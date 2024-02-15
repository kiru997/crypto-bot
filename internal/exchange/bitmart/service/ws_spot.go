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

type spotExchange struct {
	configs       *configs.AppConfig
	filterChannel map[string]struct{}
}

func NewSpotExchange(configs *configs.AppConfig) ws.Exchange {
	return &spotExchange{
		configs:       configs,
		filterChannel: helper.ArrayToMap([]string{constants.BitmartWSOpUnSubscribe}),
	}
}

func getSymbol(s string) string {
	return fmt.Sprintf("%s%s", constants.BitmartWSSpotTickerPrefix, strings.ReplaceAll(s, constants.CoinSymbolSeparateChar, constants.CoinSymbolSeparateCharUnderscore))
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSSpotBookTickerMsg{
		OP:   constants.BitmartWSOpSubscribe,
		Args: []string{getSymbol(symbol)},
	}

	time.Sleep(constants.BitmartWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSSpotBookTickerMsg{
		OP:   constants.BitmartWSOpUnSubscribe,
		Args: []string{getSymbol(symbol)},
	}

	time.Sleep(constants.BitmartWSRequestSleep)

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             enum.ExchangeTypeBitmart,
		TradingType:              enum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Bitmart.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Bitmart.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
	return s.configs.Bitmart.WSSpotBaseURL, nil
}

func (s *spotExchange) GetPingMsg() []byte {
	return []byte(`ping`)
}

func (s *spotExchange) FilterMsg(message []byte) bool {
	if string(message) == constants.BitmartWSActionPong {
		return true
	}

	action := jsoniter.Get(message, "event").ToString()
	_, skip := s.filterChannel[action]

	return skip
}
