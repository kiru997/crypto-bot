package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/exchange/gate/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"
	"github.com/gateio/gateapi-go/v6"
	"github.com/pkg/errors"
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
	configs    *configs.AppConfig
	exchange   ws.WS
	gateClient *gateapi.APIClient
}

func NewSpotService(configs *configs.AppConfig, gateClient *gateapi.APIClient) SpotService {
	return &spotService{
		configs:    configs,
		exchange:   ws.NewWS(NewSpotExchange(configs)),
		gateClient: gateClient,
	}
}

func (s *spotService) TopChange(ctx context.Context) ([]string, error) {
	res, _, err := s.gateClient.SpotApi.ListTickers(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "ListTickers")
	}

	tickers := lo.Filter(res, func(item gateapi.Ticker, _ int) bool {
		return strings.Contains(item.CurrencyPair, constants.CoinUSDT)
	})

	tickers = lo.Filter(tickers, func(item gateapi.Ticker, _ int) bool {
		vol, _ := strconv.ParseFloat(item.BaseVolume, 64)
		return vol >= s.configs.Gate.MinVol24h
	})

	sort.Slice(tickers, func(i, j int) bool {
		number, _ := strconv.ParseFloat(tickers[i].ChangePercentage, 64)
		numbeJr, _ := strconv.ParseFloat(tickers[j].ChangePercentage, 64)
		return number > numbeJr
	})

	if len(tickers) > s.configs.Gate.TopChangeLimit {
		tickers = tickers[:s.configs.Gate.TopChangeLimit]
	}

	result := make([]string, 0, len(tickers))

	for _, v := range tickers {
		result = append(result, strings.ReplaceAll(v.CurrencyPair, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar))
	}

	return result, nil
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

		var tickerMsg *dto.WSSpotTickerMessage

		err := json.Unmarshal(message, &tickerMsg)
		if err != nil {
			log.Error("ProcessTickerMsg Unmarshal error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		symbol := strings.ReplaceAll(tickerMsg.Result.CurrencyPair, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar)

		price, err := tickerMsg.Result.Last.Float64()
		if err != nil {
			log.Error("ProcessTickerMsg parse price error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		cha <- &idto.ComparePriceChanMsg{
			ExchangeType: msg.ExchangeType,
			TradingType:  msg.TradingType,
			Symbol:       symbol,
			Price:        price,
			At:           time.UnixMilli(tickerMsg.TimeMs),
			ConID:        msg.ConnID,
		}
	}
}
