package checker

import (
	"context"
	"review_reminder_bot/internal/domain/notifications"
	"review_reminder_bot/internal/domain/requests"
	"time"
)

type RepoAdapter interface {
	GetMergeRequests(ctx context.Context, groupID string, filterInterval *time.Duration) <-chan *requests.MergeRequest
	GetMergeRequestApprovals(ctx context.Context, projectID, requestID int) ([]int, error)
	GetMergeRequestParticipants(ctx context.Context, projectID, requestID int) ([]*requests.User, error)
	GetMergeRequestDiscussions(ctx context.Context, projectID, requestID int) ([]*requests.Discussion, error)
}

type StorageRepo interface {
	SaveNotification(ctx context.Context, notification *notifications.Notification) error
}

type AccessManager interface {
	IsUserAllowed(username string) bool
	IsRepositoryAllowed(repoID int) bool
	IsWebUrlAllowed(weburl string) bool
}
