package configs

import (
	"encoding/json"
	"os"

	"example.com/greetings/pkg/enum"
)

type CompareConfigExchange struct {
	Exchange        enum.ExchangeType   `json:"exchange"`
	Enable          bool                `json:"enable"`
	FutureExchanges []enum.ExchangeType `json:"future_exchanges"`
}

type CompareConfig []*CompareConfigExchange

func NewCompareConfig(configPath string) (*CompareConfig, error) {
	config := &CompareConfig{}

	rawContent, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawContent, config)
	if err != nil {
		return nil, err
	}

	filterConfig := make(CompareConfig, 0, len(*config))

	for _, v := range *config {
		filterConfig = append(filterConfig, v)
	}

	return config, nil
}
