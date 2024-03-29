package configs

type OkxConfigs struct {
	WSFutureBaseURL          string `yaml:"ws_future_base_url"`
	RefreshConnectionMinutes int    `yaml:"refresh_connection_minutes"`
	SpotMaxSubscriptions     int    `yaml:"spot_max_subscriptions"`
	FutureMaxSubscriptions   int    `yaml:"future_max_subscriptions"`
}
