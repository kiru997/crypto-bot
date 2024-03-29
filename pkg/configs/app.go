package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name     string          `yaml:"name"`
	Env      string          `yaml:"env"`
	Debug    bool            `yaml:"debug"`
	Port     string          `yaml:"port"`
	LogLevel string          `yaml:"log_level"`
	Telegram TelegramConfigs `yaml:"telegram"`
	Kucoin   KucoinConfigs   `yaml:"kucoin"`
	Mexc     MexcConfigs     `yaml:"mexc"`
	Okx      OkxConfigs      `yaml:"okx"`
	Binance  BinanceConfigs  `yaml:"binance"`
	Bybit    BybitConfigs    `yaml:"bybit"`
	Gate     GateConfigs     `yaml:"gate"`
	Bitmart  BitmartConfigs  `yaml:"bitmart"`
}

func NewConfig(configPath string) (*AppConfig, error) {
	config := &AppConfig{}

	rawContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawContent, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
