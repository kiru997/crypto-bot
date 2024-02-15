package service

import (
	"encoding/json"

	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
)

type spotExchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewSpotExchange(configs *configs.AppConfig) ws.Exchange {
	return &spotExchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{}),
	}
}

type msgArgItem struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.OkxOPSubscribe,
		"args": []*msgArgItem{{
			Channel: constants.OkxChannelIndexTicker,
			InstID:  symbol,
		}},
	}

	msg, _ := json.Marshal(data)

	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.OkxOPUnSubscribe,
		"args": []*msgArgItem{{
			Channel: constants.OkxChannelIndexTicker,
			InstID:  symbol,
		}},
	}
	msg, _ := json.Marshal(data)

	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeOkx,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Okx.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Okx.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Okx.WSFutureBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	return []byte(`ping`)
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	if string(message) == "pong" {
		return true
	}

	return false
}
