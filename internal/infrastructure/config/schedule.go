package config

type ScheduleConfig struct {
	Cron string `koanf:"cron"`
}

func NewScheduleConfig() ScheduleConfig {
	return ScheduleConfig{}
}
