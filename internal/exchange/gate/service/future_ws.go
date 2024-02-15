package service

import (
	"encoding/json"
	"strings"
	"time"

	"example.com/greetings/internal/exchange/gate/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
	jsoniter "github.com/json-iterator/go"
)

type futureExchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewFutureExchange(configs *configs.AppConfig) ws.Exchange {
	return &futureExchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{constants.GateWSChannelFuturePong, constants.GateWSEventSubscribe}),
	}
}

func getTime() int64 {
	return time.Now().Unix()
}

func getSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, constants.CoinSymbolSeparateChar, constants.CoinSymbolSeparateCharUnderscore)
}

func (*futureExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelFutureTicker,
		Event:   constants.GateWSEventSubscribe,
		Payload: []string{getSymbol(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*futureExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelFutureTicker,
		Event:   constants.GateWSEventUnSubscribe,
		Payload: []string{getSymbol(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeGateFuture,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Gate.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Gate.FutureMaxSubscriptions,
	}
}

func (s *futureExchange) GetBaseURL() (string, error) {
	return s.configs.Gate.WSFutureBaseURL, nil
}

func (s *futureExchange) GetPingMsg() []byte {
	data := &dto.WSMessage{
		Time:    getTime(),
		Channel: constants.GateWSChannelFuturePing,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "channel").ToString()
	event := jsoniter.Get(message, "event").ToString()
	_, skip := s.filterChannel[channel]
	_, skipE := s.filterChannel[event]

	return skip || skipE
}
