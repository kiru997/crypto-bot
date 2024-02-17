package service

import (
	"encoding/json"
	"fmt"
	"strings"

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
		filterChannel: helper.ArrayToMap([]string{"PONG"}),
	}
}

func getSpotTopic(s string) string {
	return fmt.Sprintf("%s%s", constants.BookerTickerParamsPrefix, strings.ReplaceAll(s, constants.CoinSymbolSeparateChar, ""))
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.MexcWSMethodSubscription,
		"params": []string{getSpotTopic(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.MexcWSMethodUnSubscription,
		"params": []string{getSpotTopic(symbol)},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeMexc,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Mexc.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Mexc.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Mexc.WSSpotBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	return []byte(`{"method":"PING"}`)
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	msg := jsoniter.Get(message, "msg").ToString()
	_, skip := s.filterChannel[msg]
	return skip || strings.Contains(msg, constants.BookerTickerParamsPrefix) || strings.Contains(msg, "no subscription success")
}
