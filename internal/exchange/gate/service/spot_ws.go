package service

import (
	"encoding/json"

	"example.com/greetings/internal/exchange/gate/dto"
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
		filterChannel: helper.ArrayToMap([]string{constants.GateWSChannelSpotPong, constants.GateWSEventSubscribe}),
	}
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelSpotTicker,
		Event:   constants.GateWSEventSubscribe,
		Payload: []string{getSymbol(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelSpotTicker,
		Event:   constants.GateWSEventUnSubscribe,
		Payload: []string{getSymbol(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeGate,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Gate.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Gate.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Gate.WSSpotBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelSpotPing,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "channel").ToString()
	event := jsoniter.Get(message, "event").ToString()
	_, skip := s.filterChannel[channel]
	_, skipE := s.filterChannel[event]

	return skip || skipE
}
