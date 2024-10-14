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

func (ln *LogNotifier) SendMessage(message string) error {
	ln.logger.Info("successfully sent", "message", message)
	return nil
}
