package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		filterChannel: helper.ArrayToMap([]string{}),
	}
}

func getParam(s string) string {
	return fmt.Sprintf("%s@%s", strings.ToLower(strings.ReplaceAll(s, constants.CoinSymbolSeperateChar, "")), "ticker")
}

func (*futureExchange) GetSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodSubcribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (*futureExchange) GetUnSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodUnSubcribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBinanceFuture,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Binance.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Binance.MaxSubscriptions,
	}
}

func (s *futureExchange) GetBaseURL() (string, error) {
	return s.configs.Binance.WSFutureBaseURL, nil
}

func (s *futureExchange) GetPingMsg() []byte {
	return []byte{}
}

func (s *futureExchange) FilterMsg(message []byte) bool {
	id := jsoniter.Get(message, "id").ToInt()
	return id != 0
}
