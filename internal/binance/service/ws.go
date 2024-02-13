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

func getParam(s string) string {
	return fmt.Sprintf("%s@%s", strings.ToLower(strings.ReplaceAll(s, constants.CoinSymbolSeperateChar, "")), "ticker")
}

func (*exchange) GetSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodSubcribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (*exchange) GetUnSubcribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"method": constants.BinanceWSMethodUnSubcribe,
		"params": []string{getParam(symbol)},
		"id":     helper.RandomNumber(13),
	}

	time.Sleep(constants.BinanceWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *exchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBinance,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Binance.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Binance.MaxSubscriptions,
	}
}

func (s *exchange) GetBaseURL() (string, error) {
	return s.configs.Binance.WSFutureBaseURL, nil
}

func (s *exchange) GetPingMsg() []byte {
	return []byte{}
}

func (s *exchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "result").ToString()
	_, skip := s.filterChannel[channel]
	return skip
}
