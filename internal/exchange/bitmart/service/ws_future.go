package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"example.com/greetings/internal/exchange/bitmart/dto"
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
		filterChannel: helper.ArrayToMap([]string{constants.BitmartWSActionSubscribe, constants.BitmartWSActionUnSubscribe}),
	}
}

func getFutureSymbol(s string) string {
	return fmt.Sprintf("%s%s", constants.BitmartWSFutureKlineBin1mPrefix, strings.ReplaceAll(s, constants.CoinSymbolSeparateChar, ""))
}

func (*futureExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSFutureTickerMsg{
		Action: constants.BitmartWSActionSubscribe,
		Args:   []string{getFutureSymbol(symbol)},
	}

	time.Sleep(constants.BitmartWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSFutureTickerMsg{
		Action: constants.BitmartWSActionUnSubscribe,
		Args:   []string{getFutureSymbol(symbol)},
	}

	time.Sleep(constants.BitmartWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBitmartFuture,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Bitmart.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Bitmart.FutureMaxSubscriptions,
	}
}

func (s *futureExchange) GetBaseURL() (string, error) {
	return s.configs.Bitmart.WSFutureBaseURL, nil
}

func (s *futureExchange) GetPingMsg() []byte {
	data := &dto.WSFutureTickerMsg{
		Action: constants.BitmartWSActionPing,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) FilterMsg(message []byte) bool {
	action := jsoniter.Get(message, "action").ToString()
	_, skip := s.filterChannel[action]

	data := jsoniter.Get(message, "data").ToString()

	return skip || strings.Contains(data, constants.BitmartWSActionPong)
}
