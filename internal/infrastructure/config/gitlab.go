package config

type GitlabConfig struct {
	Host              string `koanf:"host"`
	Token             string `koanf:"token"`
	RequestsPerSecond int    `koanf:"ratelimit"`
}

func NewGitlabConfig() GitlabConfig {
	return GitlabConfig{}
}
