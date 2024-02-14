package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"example.com/greetings/internal/kucoin/dto"
	"example.com/greetings/internal/kucoin/enum"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	ienum "example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
	kumex "github.com/Kucoin/kucoin-futures-go-sdk"
	"github.com/Kucoin/kucoin-go-sdk"
	jsoniter "github.com/json-iterator/go"
)

type kucoinFutureExchange struct {
	configs       *configs.AppConfig
	c             *kumex.ApiService
	filterChannel map[string]struct{}
}

func NewKucoinFutureExchange(configs *configs.AppConfig, c *kumex.ApiService) ws.Exchange {
	return &kucoinFutureExchange{
		configs:       configs,
		c:             c,
		filterChannel: helper.ArrayToMap([]string{"pong", "ack", "welcome"}),
	}
}

func getFutureTopic(symbol string) string {
	s := strings.ReplaceAll(strings.ReplaceAll(symbol, constants.CoinSymbolSeparateChar, ""), constants.CoinUSDT, constants.CoinUSDTM)
	return constants.KucoinFutureTopicMarketTicker + s
}

func (*kucoinFutureExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSWriteMessage{
		ID:             helper.RandomNumber(13),
		Type:           enum.WSWriteMsgTypeSubscribe,
		Topic:          getFutureTopic(symbol),
		PrivateChannel: false,
		Response:       true,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*kucoinFutureExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSWriteMessage{
		ID:             helper.RandomNumber(13),
		Type:           enum.WSWriteMsgTypeUnSubscribe,
		Topic:          getFutureTopic(symbol),
		PrivateChannel: false,
		Response:       true,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *kucoinFutureExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             ienum.ExchangeTypeKucoinFuture,
		TradingType:              ienum.TradingTypeFuture,
		RefreshConnectionMinutes: s.configs.Kucoin.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Kucoin.FutureMaxSubscriptions,
	}
}

func (s *kucoinFutureExchange) GetBaseURL() (string, error) {
	rsp, err := s.c.WebSocketPublicToken()
	if err != nil {
		return "", err
	}

	tk := &kucoin.WebSocketTokenModel{}
	if err := rsp.ReadData(tk); err != nil {
		return "", err
	}

	randomServer, err := tk.Servers.RandomServer()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s?token=%s", randomServer.Endpoint, tk.Token), nil
}

func (s *kucoinFutureExchange) GetPingMsg() []byte {
	data := &dto.WSWriteMessage{
		ID:             helper.RandomNumber(13),
		Type:           enum.WSWriteMsgTypePing,
		Topic:          "",
		PrivateChannel: false,
		Response:       false,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *kucoinFutureExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "type").ToString()
	_, skip := s.filterChannel[channel]
	return skip
}
