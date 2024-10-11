package mattermost

import (
	"log/slog"
)

type LogNotifier struct {
	logger slog.Logger
}

func NewLog() *LogNotifier {
	return &LogNotifier{
		logger: *slog.With(
			slog.String("service", "mattermost"),
		),
	}
}

func (ln *LogNotifier) SendMessage(channel string, message string) error {
	ln.logger.Info("successfully sent", "channel", channel, "message", message)
	return nil
}
