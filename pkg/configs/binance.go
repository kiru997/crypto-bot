package configs

type BinanceConfigs struct {
	WSFutureBaseURL          string `yaml:"ws_future_base_url"`
	RefreshConnectionMinutes int    `yaml:"refresh_connection_minutes"`
	MaxSubscriptions         int    `yaml:"max_subscriptions"`
}
