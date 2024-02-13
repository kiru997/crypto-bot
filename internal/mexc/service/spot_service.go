package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/mexc/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/http_client"
	"example.com/greetings/pkg/ws"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type SpotService interface {
	Subcribe(symbols []string) error
	UnSubcribe(symbols []string) error
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
	response, err := http_client.JSONRequest(ctx, fmt.Sprintf("%s%s", s.configs.Mexc.SpotAPIBaseURL, "/api/v3/ticker/24hr"), constants.HttpMethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "JSONRequest")
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("JSONRequest status code %v", response.StatusCode))
	}

	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)

	var res *dto.SpotTicker24hRes
	if err = json.Unmarshal(resBody, &res); err != nil {
		return nil, err
	}

	tickers := lo.Filter(*res, func(item *dto.SpotTicker24hResItem, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	sort.Slice(tickers, func(i, j int) bool {
		number, _ := tickers[i].PriceChangePercent.Float64()
		numbeJr, _ := tickers[j].PriceChangePercent.Float64()
		return number > numbeJr
	})

	if len(tickers) > s.configs.Mexc.TopChangeLimit {
		tickers = tickers[:s.configs.Mexc.TopChangeLimit]
	}

	result := make([]string, 0, len(tickers))

	for _, v := range tickers {
		result = append(result, strings.ReplaceAll(v.Symbol, constants.CoinUSDT, constants.CoinSymbolSeperateChar+constants.CoinUSDT))
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

func (s *spotService) Subcribe(symbols []string) error {
	return s.exchange.Subcribe(symbols)
}

func (s *spotService) UnSubcribe(symbols []string) error {
	return s.exchange.UnSubcribe(symbols)
}
