package service

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"example.com/greetings/internal/bybit/dto"
	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/ws"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	bybit "github.com/wuhewuhe/bybit.go.api"
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
	configs  *configs.AppConfig
	exchange ws.WS
}

func NewSpotService(configs *configs.AppConfig) SpotService {
	return &spotService{
		configs:  configs,
		exchange: ws.NewWS(NewSpotExchange(configs)),
	}
}

func (s *spotService) TopChange(ctx context.Context) ([]string, error) {
	client := bybit.NewBybitHttpClient("", "")
	c := client.NewMarketInfoService(map[string]interface{}{
		"category": "spot",
	})

	res, err := c.GetMarketTickers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "GetMarketTickers")
	}

	resBody, err := json.Marshal(res.Result)
	if err != nil {
		return nil, errors.Wrap(err, "Marshal")
	}

	var mtres *dto.MarketTickers

	if err = json.Unmarshal(resBody, &mtres); err != nil {
		return nil, err
	}

	tickers := lo.Filter(mtres.List, func(item *dto.MarketTickerItem, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	sort.Slice(tickers, func(i, j int) bool {
		number, _ := tickers[i].Price24hPcnt.Float64()
		numbeJr, _ := tickers[j].Price24hPcnt.Float64()
		return number > numbeJr
	})

	if len(tickers) > s.configs.Bybit.TopChangeLimit {
		tickers = tickers[:s.configs.Bybit.TopChangeLimit]
	}

	result := make([]string, 0, len(tickers))

	for _, v := range tickers {
		result = append(result, strings.ReplaceAll(v.Symbol, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT))
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
