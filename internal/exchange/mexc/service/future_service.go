package service

import (
	"encoding/json"
	"strings"
	"time"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/exchange/mexc/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"
)

type FutureService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.Msg
	GetConnections() map[string]*idto.ConnectionItem
	ProcessTickerMsg(cha chan *idto.ComparePriceChanMsg)
}

type futureService struct {
	configs  *configs.AppConfig
	exchange ws.WS
}

func NewFutureService(configs *configs.AppConfig) FutureService {
	return &futureService{
		configs:  configs,
		exchange: ws.NewWS(NewFutureExchange(configs)),
	}
}

func (s *futureService) GetConnections() map[string]*idto.ConnectionItem {
	return s.exchange.GetConnections()
}

func (s *futureService) GetMsg() chan *ws.Msg {
	return s.exchange.GetMsg()
}

func (s *futureService) RefreshConn() {
	s.exchange.RefreshConn()
}

func (s *futureService) Subscribe(symbols []string) error {
	return s.exchange.Subscribe(symbols)
}

func (s *futureService) UnSubscribe(symbols []string) error {
	return s.exchange.UnSubscribe(symbols)
}

func (s *futureService) ProcessTickerMsg(cha chan *idto.ComparePriceChanMsg) {
	for msg := range s.GetMsg() {
		message := msg.Msg

		var tickerMsg *dto.TickerMessage

		err := json.Unmarshal(message, &tickerMsg)
		if err != nil {
			log.Error("ProcessTickerMsg Unmarshal error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		symbol := strings.ReplaceAll(tickerMsg.Symbol, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar)

		cha <- &idto.ComparePriceChanMsg{
			ExchangeType: msg.ExchangeType,
			TradingType:  msg.TradingType,
			Symbol:       symbol,
			Price:        tickerMsg.Data.LastPrice,
			At:           time.UnixMilli(tickerMsg.Data.Timestamp),
			ConID:        msg.ConnID,
		}
	}
}
