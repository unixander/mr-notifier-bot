package notifier

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	domainNotifications "review_reminder_bot/internal/domain/notifications"
)

var messageTemplate *template.Template

//go:embed templates/message.gohtml
var messageTemplateFile string

func init() {
	messageTemplate = template.Must(template.New("message.gohtml").Parse(messageTemplateFile))
}

type NotificationsByTypeTemplateCtx struct {
	AwaitingReview         map[string]map[string]struct{}
	AwaitingThreadResponse map[string]map[string]struct{}
	AwaitingThreadResolve  map[string]map[string]struct{}
	AwaitingPipelineFix    map[string]map[string]struct{}
}

func NewNotificationByTypeTemplateCtx() *NotificationsByTypeTemplateCtx {
	templateCtx := &NotificationsByTypeTemplateCtx{
		AwaitingReview:         make(map[string]map[string]struct{}),
		AwaitingThreadResponse: make(map[string]map[string]struct{}),
		AwaitingThreadResolve:  make(map[string]map[string]struct{}),
		AwaitingPipelineFix:    make(map[string]map[string]struct{}),
	}

	return templateCtx
}

func (templateCtx NotificationsByTypeTemplateCtx) IsEmpty() bool {
	return (len(templateCtx.AwaitingPipelineFix) == 0 &&
		len(templateCtx.AwaitingReview) == 0 &&
		len(templateCtx.AwaitingThreadResolve) == 0 &&
		len(templateCtx.AwaitingThreadResponse) == 0)
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

func AddUser(mapStruct map[string]map[string]struct{}, notitification *domainNotifications.Notification) {
	if _, found := mapStruct[notitification.Link()]; !found {
		mapStruct[notitification.Link()] = make(map[string]struct{})
	}
	mapStruct[notitification.Link()][notitification.UserName] = struct{}{}
}

func (service *NotifierService) Run(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	notificationsByType := NewNotificationByTypeTemplateCtx()
	usernames, _ := service.StorageRepo.GetUsernamesToNotify(ctx)
	for _, username := range usernames {
		if len(username) == 0 {
			continue
		}

		notifications, err := service.StorageRepo.GetNotificationsByUsername(ctx, username)
		if err != nil {
			return err
		}

		for _, notification := range notifications {
			switch notification.Type {
			case domainNotifications.AwaitingReview:
				AddUser(notificationsByType.AwaitingReview, notification)
			case domainNotifications.AwaitingPipelineFix:
				AddUser(notificationsByType.AwaitingPipelineFix, notification)
			case domainNotifications.AwaitingThreadResolve:
				AddUser(notificationsByType.AwaitingThreadResolve, notification)
			case domainNotifications.AwaitingThreadResponse:
				AddUser(notificationsByType.AwaitingThreadResponse, notification)
			default:
				slog.Error("invalid type", "type", notification.Type)
			}
		}
	}

	if notificationsByType.IsEmpty() {
		slog.Info("nothing to send")
	} else {
		message, err := RenderMessage(notificationsByType)
		if err != nil {
			return fmt.Errorf("cannot render template: %w", err)
		}

		err = service.MessagingAdapter.SendMessage(message)
		if err != nil {
			return fmt.Errorf("message send failed: %w", err)
		}
	}

	service.StorageRepo.Clear(ctx)
	return nil
}
