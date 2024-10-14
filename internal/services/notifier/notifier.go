package notifier

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	domainNotifications "review_reminder_bot/internal/domain/notifications"

	"golang.org/x/sync/errgroup"
)

var messageTemplate *template.Template

//go:embed templates/message.gohtml
var messageTemplateFile string

func init() {
	messageTemplate = template.Must(template.New("message.gohtml").Parse(messageTemplateFile))
}

type NotificationsByTypeTemplateCtx struct {
	AwaitingReview         map[string]struct{}
	AwaitingThreadResponse map[string]struct{}
	AwaitingThreadResolve  map[string]struct{}
	AwaitingPipelineFix    map[string]struct{}
}

func NewNotificationByTypeTemplateCtx() *NotificationsByTypeTemplateCtx {
	return &NotificationsByTypeTemplateCtx{
		AwaitingReview:         make(map[string]struct{}),
		AwaitingThreadResponse: make(map[string]struct{}),
		AwaitingThreadResolve:  make(map[string]struct{}),
		AwaitingPipelineFix:    make(map[string]struct{}),
	}
}

type NotifierService struct {
	MessagingAdapter MessagingAdapter
	StorageRepo      StorageRepo
}

func New(messagingAdapter MessagingAdapter, storageRepo StorageRepo) *NotifierService {
	return &NotifierService{
		MessagingAdapter: messagingAdapter,
		StorageRepo:      storageRepo,
	}
}

func RenderMessage(templateCtx *NotificationsByTypeTemplateCtx) (string, error) {
	var buf bytes.Buffer
	err := messageTemplate.Execute(&buf, templateCtx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (service *NotifierService) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}

	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(5)

	usernames, _ := service.StorageRepo.GetUsernamesToNotify(ctx)
	for _, username := range usernames {
		if len(username) == 0 {
			continue
		}

		group.Go(func() error {
			notifications, err := service.StorageRepo.GetNotificationsByUsername(ctx, username)
			if err != nil {
				return err
			}

			notificationsByType := NewNotificationByTypeTemplateCtx()
			for _, notification := range notifications {
				switch notification.Type {
				case domainNotifications.AwaitingReview:
					notificationsByType.AwaitingReview[notification.Link()] = struct{}{}
				case domainNotifications.AwaitingPipelineFix:
					notificationsByType.AwaitingPipelineFix[notification.Link()] = struct{}{}
				case domainNotifications.AwaitingThreadResolve:
					notificationsByType.AwaitingThreadResolve[notification.Link()] = struct{}{}
				case domainNotifications.AwaitingThreadResponse:
					notificationsByType.AwaitingThreadResponse[notification.Link()] = struct{}{}
				default:
					slog.Error("invalid type", "type", notification.Type)
				}
			}

			message, err := RenderMessage(notificationsByType)
			if err != nil {
				return fmt.Errorf("cannot render template: %w", err)
			}

			userChannel := fmt.Sprintf("@%s", username)
			err = service.MessagingAdapter.SendMessage(userChannel, message)
			if err != nil {
				return fmt.Errorf("message send failed: %w", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return err
	}

	service.StorageRepo.Clear(ctx)
	return nil
}
