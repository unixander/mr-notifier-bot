package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Gitlab     GitlabConfig     `koanf:"gitlab"`
	Mattermost MattermostConfig `koanf:"mattermost"`
	Schedule   ScheduleConfig   `koanf:"schedule"`
	Settings   Settings         `koanf:"settings"`
}

func NewConfig() Config {
	config := Config{}

	config.Gitlab = NewGitlabConfig()
	config.Mattermost = NewMattermostConfig()
	config.Schedule = NewScheduleConfig()
	config.Settings = NewSettings()
	return config
}

func LoadConfig() (*Config, error) {
	koanfInstance := koanf.New(".")

	if err := koanfInstance.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		return nil, err
	}

	if err := koanfInstance.Load(env.Provider("BOT_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, "BOT_")), "_", ".", -1)
	}), nil); err != nil {
		return nil, err
	}

	config := NewConfig()
	koanfInstance.Unmarshal("", &config)
	koanfInstance.UnmarshalWithConf("settings", &config.Settings, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
	return &config, nil
}
