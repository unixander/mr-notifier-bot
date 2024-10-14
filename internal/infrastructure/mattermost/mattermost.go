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
	channel  string
	username string
}

func New(cfg *config.MattermostConfig, channel string) *MattermostAdapter {
	return &MattermostAdapter{
		logger: *slog.With(
			slog.String("service", "mattermost"),
		),
		client: &slack.Client{
			WebhookURL: cfg.IncomingWebhook,
		},
		username: cfg.BotUsername,
		channel:  channel,
	}
}

func (mm *MattermostAdapter) SendMessage(message string) error {
	if mm.channel == "" {
		return fmt.Errorf("cannot send to empty channel")
	}
	if err := mm.client.Send(&slack.Message{
		Username: mm.username,
		Text:     message,
		Channel:  mm.channel,
	}); err != nil {
		return err
	}
	mm.logger.Info("successfully sent", "message", message)
	return nil
}
