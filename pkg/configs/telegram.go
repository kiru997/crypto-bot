package configs

type TelegramConfigs struct {
	ChatID    int64  `yaml:"chat_id"`
	BotAPIKey string `yaml:"bot_api_key"`
}
