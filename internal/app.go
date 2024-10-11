package internal

import (
	"context"
	"log/slog"
	accesscontrol "review_reminder_bot/internal/infrastructure/access_control"
	"review_reminder_bot/internal/infrastructure/config"
	"review_reminder_bot/internal/infrastructure/gitlab"
	"review_reminder_bot/internal/infrastructure/mattermost"
	"review_reminder_bot/internal/infrastructure/storage/inmemory"
	"review_reminder_bot/internal/services/checker"
	"review_reminder_bot/internal/services/notifier"

	"github.com/go-co-op/gocron/v2"
)

func InitCheckerSrv(ctx context.Context, storageRepo checker.StorageRepo, cfg *config.Config) (*checker.MRCheckerService, error) {
	gitlabAdapter, err := gitlab.New(cfg.Gitlab.Host, cfg.Gitlab.Token, cfg.Gitlab.RequestsPerSecond)
	if err != nil {
		return nil, err
	}

	accessManager := accesscontrol.New(&cfg.Settings)
	checkerSvc := checker.New(
		gitlabAdapter,
		storageRepo,
		accessManager,
		cfg.Settings,
	)
	return checkerSvc, nil
}

func InitNotifier(ctx context.Context, storageRepo notifier.StorageRepo, cfg *config.Config) (*notifier.NotifierService, error) {
	mmAdapter := mattermost.New(&cfg.Mattermost)
	notificationSvc := notifier.New(
		mmAdapter,
		storageRepo,
	)
	return notificationSvc, nil
}

func Run(cfg *config.Config) error {
	ctx := context.Background()
	storageRepo := inmemory.New()

	checkerSvc, err := InitCheckerSrv(ctx, storageRepo, cfg)
	if err != nil {
		return err
	}

	notificationSvc, err := InitNotifier(ctx, storageRepo, cfg)
	if err != nil {
		return err
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	_, err = scheduler.NewJob(
		gocron.CronJob(
			cfg.Schedule.Cron,
			false,
		),
		gocron.NewTask(
			func() {
				slog.Info("check new events")
				err := checkerSvc.Run(ctx)
				if err != nil {
					slog.Error("check new events failed", "error", err)
					return
				}
				slog.Info("start sending notifications")
				err = notificationSvc.Run(ctx)
				if err != nil {
					slog.Error("send notifications failed", "error", err)
					return
				}
				slog.Info("completed")
			},
		),
	)
	if err != nil {
		return err
	}

	scheduler.Start()

	<-ctx.Done()

	return scheduler.Shutdown()
}

func RunCLI(cfg *config.Config) error {
	ctx := context.Background()
	storageRepo := inmemory.New()

	checkerSvc, err := InitCheckerSrv(ctx, storageRepo, cfg)
	if err != nil {
		return err
	}

	notificationSvc, err := InitNotifier(ctx, storageRepo, cfg)
	if err != nil {
		return err
	}

	slog.Info("check new events")
	err = checkerSvc.Run(ctx)
	if err != nil {
		slog.Error("check new events failed", "error", err)
		return err
	}
	slog.Info("start sending notifications")
	err = notificationSvc.Run(ctx)
	if err != nil {
		slog.Error("send notifications failed", "error", err)
		return err
	}
	slog.Info("completed")
	return nil
}
