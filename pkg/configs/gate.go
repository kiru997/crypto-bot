package configs

type GateConfigs struct {
	SpotAPIBaseURL           string  `yaml:"spot_api_base_url"`
	WSFutureBaseURL          string  `yaml:"ws_future_base_url"`
	WSSpotBaseURL            string  `yaml:"ws_spot_base_url"`
	RefreshConnectionMinutes int     `yaml:"refresh_connection_minutes"`
	SpotMaxSubscriptions     int     `yaml:"spot_max_subscriptions"`
	FutureMaxSubscriptions   int     `yaml:"future_max_subscriptions"`
	TopChangeLimit           int     `yaml:"top_change_limit"`
	MinVol24h                float64 `yaml:"min_vol_24h"`
}
