package service

import (
	"encoding/json"
	"fmt"

	"example.com/greetings/internal/exchange/kucoin/dto"
	"example.com/greetings/internal/exchange/kucoin/enum"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	ienum "example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/ws"
	"github.com/Kucoin/kucoin-go-sdk"
	jsoniter "github.com/json-iterator/go"
)

type spotExchange struct {
	configs       *configs.AppConfig
	c             *kucoin.ApiService
	filterChannel map[string]struct{}
}

func NewSpotExchange(configs *configs.AppConfig, c *kucoin.ApiService) ws.Exchange {
	return &spotExchange{
		configs:       configs,
		c:             c,
		filterChannel: helper.ArrayToMap([]string{"pong", "ack", "welcome"}),
	}
}

func getTopic(symbol string) string {
	return constants.KucoinTopicMarketTicker + symbol
}

func (*spotExchange) GetSubscribeMsg(symbol string) []byte {
	data := &dto.WSWriteMessage{
		ID:             helper.RandomNumber(13),
		Type:           enum.WSWriteMsgTypeSubscribe,
		Topic:          getTopic(symbol),
		PrivateChannel: false,
		Response:       true,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (*spotExchange) GetUnSubscribeMsg(symbol string) []byte {
	data := &dto.WSWriteMessage{
		ID:             helper.RandomNumber(13),
		Type:           enum.WSWriteMsgTypeUnSubscribe,
		Topic:          getTopic(symbol),
		PrivateChannel: false,
		Response:       true,
	}

	msg, _ := json.Marshal(data)
	return msg
}

func (s *spotExchange) GetConfig() *ws.ExChangeConfig {
	return &ws.ExChangeConfig{
		ExchangeType:             ienum.ExchangeTypeKucoin,
		TradingType:              ienum.TradingTypeSpot,
		RefreshConnectionMinutes: s.configs.Kucoin.RefreshConnectionMinutes,
		MaxSubscriptions:         s.configs.Kucoin.SpotMaxSubscriptions,
	}
}

func (s *spotExchange) GetBaseURL() (string, error) {
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

func (s *spotExchange) GetPingMsg() []byte {
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

func (s *spotExchange) FilterMsg(message []byte) bool {
	channel := jsoniter.Get(message, "type").ToString()
	_, skip := s.filterChannel[channel]
	return skip
}
