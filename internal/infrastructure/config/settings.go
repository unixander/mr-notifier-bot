package config

import (
	"time"
)

type Settings struct {
	AllowedUsers                []string       `koanf:"users.allowed"`
	IgnoredUsers                []string       `koanf:"users.ignored"`
	AllowedRepositories         []int          `koanf:"repositories.allowed"`
	IgnoredRepositories         []int          `koanf:"repositories.ignored"`
	ApprovalsRequired           int            `koanf:"approvals.count"`
	MergeRequestsFilterInterval *time.Duration `koanf:"filter.interval"`
	GroupID                     string         `koanf:"group"`
	IgnoredWebUrlsRegexp        []string       `koanf:"weburls.ignored"`
	NotificationChannel         string         `koanf:"notifications.channel"`
}

func NewSettings() Settings {
	return Settings{}
}
