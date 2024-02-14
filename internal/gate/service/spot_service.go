package service

import (
	"context"
	"sort"
	"strconv"
	"strings"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/ws"
	"github.com/gateio/gateapi-go/v6"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type SpotService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.MsgChan
	GetConnections() map[string]*idto.ConnectionItem
	TopChange(context.Context) ([]string, error)
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

func (s *spotService) GetMsg() chan *ws.MsgChan {
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
