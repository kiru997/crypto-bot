package configs

type BinanceConfigs struct {
	SpotAPIBaseURL           string `yaml:"spot_api_base_url"`
	WSSpotBaseURL            string `yaml:"ws_spot_base_url"`
	WSFutureBaseURL          string `yaml:"ws_future_base_url"`
	RefreshConnectionMinutes int    `yaml:"refresh_connection_minutes"`
	MaxSubscriptions         int    `yaml:"max_subscriptions"`
	TopChangeLimit           int    `yaml:"top_change_limit"`
}
