package service

import (
	"encoding/json"
	"strings"

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
		filterChannel: helper.ArrayToMap([]string{"pong", "rs.sub.ticker"}),
	}
}

func (*futureExchange) GetSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.MexcWSMethodSubTicker,
		"param": map[string]string{
			"symbol": strings.ReplaceAll(symbol, constants.CoinSymbolSeperateChar, "_"),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*futureExchange) GetUnSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.MexcWSMethodUnSubTicker,
		"param": map[string]string{
			"symbol": strings.ReplaceAll(symbol, constants.CoinSymbolSeperateChar, "_"),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeMexcFuture,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Mexc.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Mexc.MaxSubscriptions,
	}
}

func (s *futureExchange) GetBaseURL() (string, error) {
	return s.configs.Mexc.WSFutureBaseURL, nil
}

func (s *futureExchange) GetPingMsg() []byte {
	return []byte(`{"method":"ping"}`)
}

func (s *futureExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "channel").ToString()
	_, skip := s.filterChannel[channel]
	
	notExist := false
	if channel == "rs.error" {
		dataMsg := jsoniter.Get(message, "data").ToString()
		notExist = strings.Contains(dataMsg, "not exists")
	}

	return skip || notExist
}
