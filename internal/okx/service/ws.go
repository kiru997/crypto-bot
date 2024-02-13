package service

import (
	"encoding/json"

	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
)

type exchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewExchange(configs *configs.AppConfig) ws.Exchange {
	return &exchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{}),
	}
}

type msgArgItem struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

func (*exchange) GetSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.OkxOPSubcribe,
		"args": []*msgArgItem{{
			Channel: constants.OkxChannelIndexTicker,
			InstId:  symbol,
		}},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*exchange) GetUnSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.OkxOPUnSubcribe,
		"args": []*msgArgItem{{
			Channel: constants.OkxChannelIndexTicker,
			InstId:  symbol,
		}},
	}
	msg, _ := json.Marshal(data)
	return msg
}

func (s *exchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeOkx,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Okx.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Okx.MaxSubscriptions,
	}
}

func (s *exchange) GetBaseURL() (string, error) {
	return s.configs.Okx.WSFutureBaseURL, nil
}

func (s *exchange) GetPingMsg() []byte {
	return []byte(`ping`)
}

func (s *exchange) FilterMsg(message []byte) bool {
	if string(message) == "pong" {
		return true
	}

	return false
}
