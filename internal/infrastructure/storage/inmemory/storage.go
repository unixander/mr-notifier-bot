package inmemory

import (
	"context"
	"review_reminder_bot/internal/domain/notifications"
)

type InMemoryRepo struct {
	notificationsTable *NotificationsTable
}

func New() *InMemoryRepo {
	return &InMemoryRepo{
		notificationsTable: NewNotificationsTable(),
	}
}

func (repo *InMemoryRepo) Clear(ctx context.Context) error {
	repo.notificationsTable.Clear()
	return nil
}

func (repo *InMemoryRepo) SaveNotification(ctx context.Context, notification *notifications.Notification) error {
	repo.notificationsTable.SaveNotification(notification)
	return nil
}

func (repo *InMemoryRepo) GetUsernamesToNotify(ctx context.Context) ([]string, error) {
	usernames := repo.notificationsTable.GetUsernamesToNotify()
	return usernames, nil
}

func (repo *InMemoryRepo) GetNotificationsByUsername(ctx context.Context, username string) ([]*notifications.Notification, error) {
	return repo.notificationsTable.GetNotificationsByUsername(username)
}
