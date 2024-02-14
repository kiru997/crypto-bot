package service

import (
	"context"
	"sort"
	"strings"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/ws"
	"github.com/Kucoin/kucoin-go-sdk"
	"github.com/samber/lo"
)

type KucoinSpotService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.MsgChan
	GetConnections() map[string]*idto.ConnectionItem
	TopChange(context.Context) ([]string, error)
}

type kucoinSpotService struct {
	configs  *configs.AppConfig
	c        *kucoin.ApiService
	exchange ws.WS
}

func NewKucoinSpotService(configs *configs.AppConfig, c *kucoin.ApiService) KucoinSpotService {
	return &kucoinSpotService{
		configs:  configs,
		c:        c,
		exchange: ws.NewWS(NewKucoinExchange(configs, c)),
	}
}

func (s *kucoinSpotService) TopChange(_ context.Context) ([]string, error) {
	resp, err := s.c.Tickers()
	if err != nil {
		return nil, err
	}

	os := &kucoin.TickersResponseModel{}
	resp.ReadData(&os)

	tickers := lo.Filter(os.Tickers, func(item *kucoin.TickerModel, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	sort.Slice(tickers, func(i, j int) bool {
		return tickers[i].ChangeRate > tickers[j].ChangeRate
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

func (s *kucoinSpotService) GetConnections() map[string]*idto.ConnectionItem {
	return s.exchange.GetConnections()
}

func (s *kucoinSpotService) GetMsg() chan *ws.MsgChan {
	return s.exchange.GetMsg()
}

func (s *kucoinSpotService) RefreshConn() {
	s.exchange.RefreshConn()
}

func (s *kucoinSpotService) Subscribe(symbols []string) error {
	return s.exchange.Subscribe(symbols)
}

func (s *kucoinSpotService) UnSubscribe(symbols []string) error {
	return s.exchange.UnSubscribe(symbols)
}
