package checker

import (
	"context"
	"log/slog"
	domainNotifications "review_reminder_bot/internal/domain/notifications"
	domainRequests "review_reminder_bot/internal/domain/requests"
	"review_reminder_bot/internal/infrastructure/config"

	"golang.org/x/sync/errgroup"
)

type MRCheckerService struct {
	RepoAdapter   RepoAdapter
	Settings      config.Settings
	Storage       StorageRepo
	AccessManager AccessManager
}

func New(repo RepoAdapter, storageRepo StorageRepo, accessManager AccessManager, settings config.Settings) *MRCheckerService {
	return &MRCheckerService{
		RepoAdapter:   repo,
		Settings:      settings,
		Storage:       storageRepo,
		AccessManager: accessManager,
	}
}

func (service *MRCheckerService) Notify(ctx context.Context, notificationChan chan *domainNotifications.Notification, notification *domainNotifications.Notification) {
	if !service.AccessManager.IsUserAllowed(notification.UserName) {
		slog.Debug("notifications not allowed for user", "user", notification.UserName)
		return
	}

	select {
	case <-ctx.Done():
	case notificationChan <- notification:
	}
}

func (service *MRCheckerService) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	requestsChan := service.RepoAdapter.GetMergeRequests(ctx, service.Settings.GroupID, service.Settings.MergeRequestsFilterInterval)
	workerGroup, _ := errgroup.WithContext(ctx)
	workerGroup.SetLimit(5)

	resultChan := make(chan *domainNotifications.Notification)

	go func() {
	RESULT_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			case notification, ok := <-resultChan:
				if !ok {
					break RESULT_LOOP
				}
				if err := service.Storage.SaveNotification(ctx, notification); err != nil {
					slog.Error("error saving notification", "error", err)
				}
			}
		}

	}()

	for request := range requestsChan {
		if !service.AccessManager.IsRepositoryAllowed(request.ID) || !service.AccessManager.IsRepositoryAllowed(request.IID) {
			slog.Info("skipped repository", "repoID", request.ID)
			continue
		}
		if !service.AccessManager.IsWebUrlAllowed(request.WebURL) {
			continue
		}

		workerGroup.Go(func() error {
			slog.Info("start checking merge request", "request", request.WebURL)
			defer slog.Info("finished processing merge request", "request", request.WebURL)

			unresolvedParticipants := service.checkUnresolvedDiscussions(ctx, request, resultChan)
			service.checkApprovals(ctx, request, unresolvedParticipants, resultChan)

			if request.Pipeline != nil && request.Pipeline.Status == domainRequests.PipelineFailed {
				service.Notify(ctx, resultChan, &domainNotifications.Notification{
					UserName:  request.Author.Username,
					RequestID: request.ID,
					ProjectID: request.ProjectID,
					Type:      domainNotifications.AwaitingPipelineFix,
					WebURL:    request.WebURL,
				})
			}

			return nil
		})
	}

	go func() {
		workerGroup.Wait()
		close(resultChan)
	}()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return workerGroup.Wait()
}
