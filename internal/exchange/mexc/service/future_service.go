package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/exchange/mexc/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/http_client"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"
	"github.com/pkg/errors"
	"github.com/samber/lo"
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

func (s *futureService) TopChange(ctx context.Context) ([]string, error) {
	response, err := http_client.JSONRequest(ctx, fmt.Sprintf("%s%s", s.configs.Mexc.FutureAPIBaseURL, "/api/v1/contract/ticker"), constants.HttpMethodGet, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "JSONRequest")
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("JSONRequest status code %v", response.StatusCode))
	}

	defer response.Body.Close()

	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	var res *dto.FutureTickerResponse
	if err = json.Unmarshal(resBody, &res); err != nil {
		return nil, err
	}

	tickers := lo.Filter(res.Data, func(item *dto.FutureTickerResponseData, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	tickers = lo.Filter(tickers, func(item *dto.FutureTickerResponseData, _ int) bool {
		return item.Volume24 >= s.configs.Mexc.MinVol24h
	})

	sort.Slice(tickers, func(i, j int) bool {
		return math.Abs(tickers[i].RiseFallRate) > math.Abs(tickers[j].RiseFallRate)
	})

	if len(tickers) > s.configs.Mexc.FutureTopChangeLimit {
		tickers = tickers[:s.configs.Mexc.FutureTopChangeLimit]
	}

	result := make([]string, 0, len(tickers))

	for _, v := range tickers {
		result = append(result, strings.ReplaceAll(v.Symbol, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar))
	}

	return result, nil
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
