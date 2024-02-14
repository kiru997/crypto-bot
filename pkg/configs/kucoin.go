package configs

type KucoinConfigs struct {
	SpotAPIBaseURL           string `yaml:"spot_api_base_url"`
	FutureAPIBaseURL         string `yaml:"future_api_base_url"`
	RefreshConnectionMinutes int    `yaml:"refresh_connection_minutes"`
	SpotMaxSubscriptions     int    `yaml:"spot_max_subscriptions"`
	FutureMaxSubscriptions   int    `yaml:"future_max_subscriptions"`
	TopChangeLimit           int    `yaml:"top_change_limit"`
}
