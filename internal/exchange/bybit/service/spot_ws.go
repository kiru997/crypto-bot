package service

import (
	"encoding/json"

	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
	jsoniter "github.com/json-iterator/go"
)

type spotExchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewSpotExchange(configs *configs.AppConfig) ws.Exchange {
	return &spotExchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{"ping", constants.BybitWSMethodSubscription, constants.BybitWSMethodUnSubscription}),
	}
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.BybitWSMethodSubscription,
		"args": []string{
			getSymbol(symbol),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.BybitWSMethodUnSubscription,
		"args": []string{
			getSymbol(symbol),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBybit,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Bybit.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Bybit.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Bybit.WSSpotBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	return []byte(`{"op":"ping"}`)
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "op").ToString()
	_, skip := s.filterChannel[channel]
	return skip
}
