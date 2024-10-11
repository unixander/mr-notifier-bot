package main

import (
	"log/slog"
	"os"
	"review_reminder_bot/internal"
	configPkg "review_reminder_bot/internal/infrastructure/config"
	"review_reminder_bot/internal/infrastructure/logger"
)

var (
	Version string = "develop"
)

func main() {
	logger.Setup()
	slog.Info("service info", "version", Version)
	config, err := configPkg.LoadConfig()
	if err != nil {
		slog.Error("cannot load config", "error", err)
		os.Exit(1)
	}
	slog.Info("config loaded")
	if err := internal.Run(config); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
	os.Exit(0)
}
