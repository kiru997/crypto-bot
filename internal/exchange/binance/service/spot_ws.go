package service

import (
	"encoding/json"
	"time"

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
		filterChannel: helper.ArrayToMap([]string{}),
	}
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodSubscribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodUnSubscribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBinance,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Binance.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Binance.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Binance.WSSpotBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	return []byte{}
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	id := jsoniter.Get(message, "id").ToInt()
	return id != 0
}
