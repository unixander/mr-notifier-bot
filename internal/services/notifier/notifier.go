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
	AwaitingReview         []*domainNotifications.Notification
	AwaitingThreadResponse []*domainNotifications.Notification
	AwaitingThreadResolve  []*domainNotifications.Notification
	AwaitingPipelineFix    []*domainNotifications.Notification
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

			notificationsByType := &NotificationsByTypeTemplateCtx{}
			for _, notification := range notifications {
				switch notification.Type {
				case domainNotifications.AwaitingReview:
					notificationsByType.AwaitingReview = append(notificationsByType.AwaitingReview, notification)
				case domainNotifications.AwaitingPipelineFix:
					notificationsByType.AwaitingPipelineFix = append(notificationsByType.AwaitingPipelineFix, notification)
				case domainNotifications.AwaitingThreadResolve:
					notificationsByType.AwaitingThreadResolve = append(notificationsByType.AwaitingThreadResolve, notification)
				case domainNotifications.AwaitingThreadResponse:
					notificationsByType.AwaitingThreadResponse = append(notificationsByType.AwaitingThreadResponse, notification)
				default:
					slog.Error("invalid type", "type", notification.Type)
				}
			}

			message, err := RenderMessage(notificationsByType)
			if err != nil {
				slog.Error("cannot render template", "error", err.Error())
				return nil
			}

			userChannel := fmt.Sprintf("@%s", username)
			err = service.MessagingAdapter.SendMessage(userChannel, message)
			if err != nil {
				slog.Error("message send failed", "error", err.Error())
				return nil
			}
			return nil
		})
	}
	group.Wait()

	service.StorageRepo.Clear(ctx)
	return nil
}
