package notifier

import (
	"context"
	"review_reminder_bot/internal/domain/notifications"
)

type MessagingAdapter interface {
	SendMessage(string) error
}

type StorageRepo interface {
	Clear(ctx context.Context) error
	GetUsernamesToNotify(ctx context.Context) ([]string, error)
	GetNotificationsByUsername(ctx context.Context, username string) ([]*notifications.Notification, error)
}
