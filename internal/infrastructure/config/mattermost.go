package config

type MattermostConfig struct {
	IncomingWebhook string `koanf:"webhook"`
	BotUsername     string `koanf:"username"`
}

func NewMattermostConfig() MattermostConfig {
	return MattermostConfig{}
}
