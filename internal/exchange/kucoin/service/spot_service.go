package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/exchange/kucoin/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"
	"github.com/Kucoin/kucoin-go-sdk"
	"github.com/samber/lo"
)

type SpotService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.Msg
	GetConnections() map[string]*idto.ConnectionItem
	TopChange(context.Context) ([]string, error)
	ProcessTickerMsg(cha chan *idto.ComparePriceChanMsg)
}

type spotService struct {
	configs  *configs.AppConfig
	c        *kucoin.ApiService
	exchange ws.WS
}

func NewSpotService(configs *configs.AppConfig, c *kucoin.ApiService) SpotService {
	return &spotService{
		configs:  configs,
		c:        c,
		exchange: ws.NewWS(NewSpotExchange(configs, c)),
	}
}

func (s *spotService) TopChange(_ context.Context) ([]string, error) {
	resp, err := s.c.Tickers()
	if err != nil {
		return nil, err
	}

	os := &kucoin.TickersResponseModel{}
	resp.ReadData(&os)

	tickers := lo.Filter(os.Tickers, func(item *kucoin.TickerModel, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	tickers = lo.Filter(tickers, func(item *kucoin.TickerModel, _ int) bool {
		vol, _ := strconv.ParseFloat(item.VolValue, 64)
		return vol >= s.configs.Kucoin.MinVol24h
	})

	sort.Slice(tickers, func(i, j int) bool {
		number, _ := strconv.ParseFloat(tickers[i].ChangeRate, 64)
		numbeJr, _ := strconv.ParseFloat(tickers[j].ChangeRate, 64)
		return number > numbeJr
	})

	if len(tickers) > s.configs.Kucoin.TopChangeLimit {
		tickers = tickers[:s.configs.Kucoin.TopChangeLimit]
	}

	res := make([]string, 0, len(tickers))

	for _, v := range tickers {
		res = append(res, v.Symbol)
	}

	return res, nil
}

func (s *spotService) GetConnections() map[string]*idto.ConnectionItem {
	return s.exchange.GetConnections()
}

func (s *spotService) GetMsg() chan *ws.Msg {
	return s.exchange.GetMsg()
}

func (s *spotService) RefreshConn() {
	s.exchange.RefreshConn()
}

func (s *spotService) Subscribe(symbols []string) error {
	return s.exchange.Subscribe(symbols)
}

func (s *spotService) UnSubscribe(symbols []string) error {
	return s.exchange.UnSubscribe(symbols)
}

func (s *spotService) ProcessTickerMsg(cha chan *idto.ComparePriceChanMsg) {
	for msg := range s.GetMsg() {
		message := msg.Msg

		var tickerMsg *dto.MarketTicker

		err := json.Unmarshal(message, &tickerMsg)
		if err != nil {
			log.Error("ProcessTickerMsg Unmarshal error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		symbol := strings.ReplaceAll(tickerMsg.Topic, constants.KucoinTopicMarketTicker, "")

		price, err := tickerMsg.Data.Price.Float64()
		if err != nil {
			log.Error("ProcessTickerMsg parse price error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		cha <- &idto.ComparePriceChanMsg{
			Symbol:       symbol,
			Price:        price,
			ExchangeType: msg.ExchangeType,
			At:           time.UnixMilli(tickerMsg.Data.Time),
			TradingType:  msg.TradingType,
			ConID:        msg.ConnID,
		}
	}
}
