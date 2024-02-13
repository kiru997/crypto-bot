package configs

type KucoinConfigs struct {
	SpotAPIBaseURL           string `yaml:"spot_api_base_url"`
	FutureAPIBaseURL         string `yaml:"future_api_base_url"`
	RefreshConnectionMinutes int    `yaml:"refresh_connection_minutes"`
	MaxSubscriptions         int    `yaml:"max_subscriptions"`
	TopChangeLimit           int    `yaml:"top_change_limit"`
}
