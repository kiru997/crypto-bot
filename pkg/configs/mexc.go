package configs

type MexcConfigs struct {
	SpotAPIBaseURL           string  `yaml:"spot_api_base_url"`
	FutureAPIBaseURL         string  `yaml:"future_api_base_url"`
	WSFutureBaseURL          string  `yaml:"ws_future_base_url"`
	WSSpotBaseURL            string  `yaml:"ws_spot_base_url"`
	RefreshConnectionMinutes int     `yaml:"refresh_connection_minutes"`
	SpotMaxSubscriptions     int     `yaml:"spot_max_subscriptions"`
	FutureMaxSubscriptions   int     `yaml:"future_max_subscriptions"`
	TopChangeLimit           int     `yaml:"top_change_limit"`
	FutureTopChangeLimit     int     `yaml:"future_top_change_limit"`
	SpotMinVol24h            float64 `yaml:"spot_min_vol_24h"`
}
