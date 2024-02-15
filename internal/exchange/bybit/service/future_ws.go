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

type futureExchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewFutureExchange(configs *configs.AppConfig) ws.Exchange {
	return &futureExchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{"ping", constants.BybitWSMethodSubscription, constants.BybitWSMethodUnSubscription}),
	}
}

func getSymbol(symbol string) string {
	return fmt.Sprintf("%s%s", constants.BybitTickerParamsPrefix, strings.ReplaceAll(symbol, constants.CoinSymbolSeparateChar+constants.CoinUSDT, constants.CoinUSDT))
}

func (*futureExchange) GetSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.BybitWSMethodSubscription,
		"args": []string{
			getSymbol(symbol),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*futureExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := map[string]interface{}{
		"op": constants.BybitWSMethodUnSubscription,
		"args": []string{
			getSymbol(symbol),
		},
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *futureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBybitFuture,
		TradingType:              enum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Bybit.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Bybit.FutureMaxSubscriptions,
	}
}

func (s *futureExchange) GetBaseURL() (string, error) {
	return s.configs.Bybit.WSFutureBaseURL, nil
}

func (s *futureExchange) GetPingMsg() []byte {
	return []byte(`{"op":"ping"}`)
}

func (s *futureExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "op").ToString()
	_, skip := s.filterChannel[channel]
	return skip
}
