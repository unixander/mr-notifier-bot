package checker

import (
	"context"
	"log/slog"
	domainNotifications "review_reminder_bot/internal/domain/notifications"
	domainRequests "review_reminder_bot/internal/domain/requests"
)

func (service *MRCheckerService) checkApprovals(
	ctx context.Context,
	request *domainRequests.MergeRequest,
	notificationChan chan *domainNotifications.Notification,
) {
	// Check approvals
	approvedBy, err := service.RepoAdapter.GetMergeRequestApprovals(ctx, request.ProjectID, request.IID)
	if err != nil {
		slog.Error("cannot get approvals", "error", err)
		return
	}

	if len(approvedBy) >= service.Settings.ApprovalsRequired {
		return
	}

	// Prepair HashSet to check users
	approvedBySet := make(map[int]struct{}, len(approvedBy))
	for _, approvedBy := range approvedBy {
		approvedBySet[approvedBy] = struct{}{}
	}

	// Get participants except Author and approvedBy
	participants, err := service.RepoAdapter.GetMergeRequestParticipants(ctx, request.ProjectID, request.IID)
	if err != nil {
		slog.Error("cannot get participants", "error", err)
		return
	}
	for _, participant := range participants {
		if ctx.Err() != nil {
			return
		}
		if _, ok := approvedBySet[participant.ID]; ok {
			continue
		}
		if participant.ID == request.Author.ID || request.IsAssignee(participant.ID) {
			continue
		}

		service.Notify(ctx, notificationChan, &domainNotifications.Notification{
			UserName:  participant.Username,
			ProjectID: request.ProjectID,
			RequestID: request.ID,
			Type:      domainNotifications.AwaitingReview,
			WebURL:    request.WebURL,
		})
	}
}
