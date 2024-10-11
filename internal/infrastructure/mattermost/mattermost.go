package mattermost

import (
	"fmt"
	"log/slog"
	"review_reminder_bot/internal/infrastructure/config"

	"github.com/int128/slack"
)

type MattermostAdapter struct {
	logger   slog.Logger
	client   *slack.Client
	username string
}

func New(cfg *config.MattermostConfig) *MattermostAdapter {
	return &MattermostAdapter{
		logger: *slog.With(
			slog.String("service", "mattermost"),
		),
		client: &slack.Client{
			WebhookURL: cfg.IncomingWebhook,
		},
		username: cfg.BotUsername,
	}
}

func (mm *MattermostAdapter) SendMessage(channel string, message string) error {
	if channel == "" {
		return fmt.Errorf("cannot send to empty channel")
	}
	if err := mm.client.Send(&slack.Message{
		Username: mm.username,
		Text:     message,
		Channel:  channel,
	}); err != nil {
		return err
	}
	mm.logger.Info("successfully sent", "channel", channel, "message", message)
	return nil
}
