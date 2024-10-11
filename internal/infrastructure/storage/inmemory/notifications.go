package inmemory

import (
	"fmt"
	"maps"
	"review_reminder_bot/internal/domain/notifications"
	"sync"
)

type NotificationsTable struct {
	mu                *sync.RWMutex
	storageByUsername map[string][]*notifications.Notification
}

func NewNotificationsTable() *NotificationsTable {
	return &NotificationsTable{
		mu:                &sync.RWMutex{},
		storageByUsername: make(map[string][]*notifications.Notification),
	}
}

func (nt *NotificationsTable) SaveNotification(notification *notifications.Notification) error {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	nt.storageByUsername[notification.UserName] = append(nt.storageByUsername[notification.UserName], notification)
	return nil
}

func (nt *NotificationsTable) GetUsernamesToNotify() []string {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	usernames := make([]string, 0, len(nt.storageByUsername))
	for username := range maps.Keys(nt.storageByUsername) {
		usernames = append(usernames, username)
	}
	return usernames
}

func (nt *NotificationsTable) GetNotificationsByUsername(username string) ([]*notifications.Notification, error) {
	nt.mu.RLock()
	defer nt.mu.RUnlock()

	if notifications, ok := nt.storageByUsername[username]; ok {
		return notifications, nil
	}
	return nil, fmt.Errorf("notifications not found")
}

func (nt *NotificationsTable) Clear() {
	nt.mu.Lock()
	defer nt.mu.Unlock()

	nt.storageByUsername = make(map[string][]*notifications.Notification)
}
