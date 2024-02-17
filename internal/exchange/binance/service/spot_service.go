package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	idto "example.com/greetings/internal/dto"
	"example.com/greetings/internal/exchange/binance/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/http_client"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type SpotService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.Msg
	GetConnections() map[string]*idto.ConnectionItem
	TopChange(ctx context.Context) ([]string, error)
	ProcessTickerMsg(cha chan *idto.ComparePriceChanMsg)
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
	response, err := http_client.JSONRequest(ctx, fmt.Sprintf("%s%s", s.configs.Binance.SpotAPIBaseURL, "/api/v3/ticker/24hr"), constants.HttpMethodGet, nil, nil)
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

	tickers := lo.Filter(*res, func(item *dto.SpotTicker24h, _ int) bool {
		return strings.Contains(item.Symbol, constants.CoinUSDT)
	})

	tickers = lo.Filter(tickers, func(item *dto.SpotTicker24h, _ int) bool {
		vol, _ := item.QuoteVolume.Float64()
		return vol >= s.configs.Binance.SpotMinVol24h
	})

	sort.Slice(tickers, func(i, j int) bool {
		number, _ := tickers[i].PriceChangePercent.Float64()
		numbeJr, _ := tickers[j].PriceChangePercent.Float64()

		return number > numbeJr
	})

	if len(tickers) > s.configs.Binance.TopChangeLimit {
		tickers = tickers[:s.configs.Binance.TopChangeLimit]
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

		var tickerMsg *dto.TickerMsg

		err := json.Unmarshal(message, &tickerMsg)
		if err != nil {
			log.Error("ProcessTickerMsg Unmarshal error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		symbol := strings.ReplaceAll(tickerMsg.Symbol, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

		price, err := tickerMsg.LastPrice.Float64()
		if err != nil {
			log.Error("ProcessTickerMsg parse price error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		vol, err := tickerMsg.QuoteVolume.Float64()
		if err != nil {
			log.Error("ProcessTickerMsg parse vol error",
				log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]),
				log.String("tradingType", enum.TradingTypeName[msg.TradingType]),
				log.Any("error", err), log.ByteString("msg", message))

			continue
		}

		if vol < s.configs.Binance.SpotMinVol24h {
			continue
		}

		cha <- &idto.ComparePriceChanMsg{
			ExchangeType: msg.ExchangeType,
			TradingType:  msg.TradingType,
			Symbol:       symbol,
			Price:        price,
			At:           time.UnixMilli(tickerMsg.CloseTime),
			ConID:        msg.ConnID,
		}
	}
}
