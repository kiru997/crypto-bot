package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	bdto "example.com/greetings/internal/binance/dto"
	"example.com/greetings/internal/dto"
	kdto "example.com/greetings/internal/kucoin/dto"
	mdto "example.com/greetings/internal/mexc/dto"

	okxdto "example.com/greetings/internal/okx/dto"

	bs "example.com/greetings/internal/binance/service"
	ks "example.com/greetings/internal/kucoin/service"

	ms "example.com/greetings/internal/mexc/service"
	okxs "example.com/greetings/internal/okx/service"

	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/helper"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/ws"

	"github.com/avast/retry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/samber/lo"
)

type CompareService interface {
	WatchTopChange()
	SendNoti()
	RefreshConn()
	GetCurrentSymbolItems() []*dto.CompareSymbolNotiItem
	GetConnections() map[string]*dto.ConnectionItem
}

type ExchangService interface {
	TopChange(context.Context) ([]string, error)
}

type compareService struct {
	configs            *configs.AppConfig
	compareConfigs     *configs.CompareConfig
	mapSymbolItem      map[string]*dto.CompareSymbolNotiItem
	priceItemChan      chan *dto.ComparePriceChanMsg
	mapExchanges       map[enum.ExchangeType]ws.WS
	mapExchangeService map[enum.ExchangeType]ExchangService
}

func NewCompareService(configs *configs.AppConfig,
	compareConfigs *configs.CompareConfig,
	kucoinSpotService ks.KucoinSpotService,
	kucoinFutureService ks.KucoinFutureService,
	binanceService bs.SpotService,
	binanceFutureService bs.FutureService,
	mexcFutureService ms.FutureService,
	mexcSpotService ms.SpotService,
	okxFutureService okxs.FutureService) CompareService {
	return &compareService{
		configs:        configs,
		compareConfigs: compareConfigs,
		mapSymbolItem:  map[string]*dto.CompareSymbolNotiItem{},
		priceItemChan:  make(chan *dto.ComparePriceChanMsg),
		mapExchanges: map[enum.ExchangeType]ws.WS{
			enum.ExchangeTypeKucoin:        kucoinSpotService,
			enum.ExchangeTypeMexc:          mexcSpotService,
			enum.ExchangeTypeMexcFuture:    mexcFutureService,
			enum.ExchangeTypeOkx:           okxFutureService,
			enum.ExchangeTypeKucoinFuture:  kucoinFutureService,
			enum.ExchangeTypeBinance:       binanceService,
			enum.ExchangeTypeBinanceFuture: binanceFutureService,
		},
		mapExchangeService: map[enum.ExchangeType]ExchangService{
			enum.ExchangeTypeKucoin:  kucoinSpotService,
			enum.ExchangeTypeMexc:    mexcSpotService,
			enum.ExchangeTypeBinance: binanceService,
		},
	}
}

func (s *compareService) GetBaseCompareExchange() []enum.ExchangeType {
	res := []enum.ExchangeType{}

	for _, v := range *s.compareConfigs {
		if !v.Enable {
			continue
		}
		res = append(res, v.Exchange)
	}

	res = lo.Uniq(res)

	return res
}

func (s *compareService) GetAvailExchange() []enum.ExchangeType {
	res := []enum.ExchangeType{}

	for _, v := range *s.compareConfigs {
		if !v.Enable {
			continue
		}
		res = append(res, v.Exchange)
		res = append(res, v.FutureExchanges...)
	}

	res = lo.Uniq(res)

	return res
}

func (s *compareService) RefreshConn() {
	log.Info("RefreshConn start")

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapExchanges[v]
		if !ok {
			log.Error("compareService RefreshConn mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}
		ex.RefreshConn()
	}
}

func (s *compareService) GetCurrentSymbolItems() []*dto.CompareSymbolNotiItem {
	result := make([]*dto.CompareSymbolNotiItem, 0, len(s.mapSymbolItem))

	for _, v := range s.mapSymbolItem {
		result = append(result, v)
	}

	return result
}

func (s *compareService) GetConnections() map[string]*dto.ConnectionItem {
	newMap := map[string]*dto.ConnectionItem{}

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapExchanges[v]
		if !ok {
			log.Error("compareService GetConnections mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}
		mc := ex.GetConnections()
		for k, i := range mc {
			newMap[k] = i
		}
	}

	return newMap
}

func (s *compareService) WatchTopChange() {
	log.Info("WatchTopChange start")

	symbols := []string{}

	for _, v := range s.GetBaseCompareExchange() {
		ex, ok := s.mapExchangeService[v]
		if !ok {
			log.Error("compareService WatchTopChange mapExchangeService not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

		defer cancel()

		s, err := ex.TopChange(ctx)
		if err != nil {
			log.Error("compareService WatchTopChange TopChange error", log.Any("error", err), log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		symbols = append(symbols, s...)
	}

	symbols = lo.Uniq(symbols)

	symbolSubcribes := make([]string, 0, len(s.mapSymbolItem))

	for _, v := range s.mapSymbolItem {
		symbolSubcribes = append(symbolSubcribes, v.Symbol)
	}

	add, deleteS := lo.Difference(symbols, symbolSubcribes)

	for _, v := range add {
		s.mapSymbolItem[v] = &dto.CompareSymbolNotiItem{
			Symbol: v,
		}
	}

	for _, v := range deleteS {
		delete(s.mapSymbolItem, v)
	}

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapExchanges[v]
		if !ok {
			log.Error("compareService GetConnections mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		err := ex.Subcribe(add)
		if err != nil {
			log.Error("TopChange kucoinSpotService Subcribe error", log.Any("error", err), log.Any("symbols", add), log.String("exchange", enum.ExchangeTypeName[v]))
			return
		}

		err = ex.UnSubcribe(deleteS)
		if err != nil {
			log.Error("TopChange kucoinSpotService UnSubcribe error", log.Any("error", err), log.Any("symbols", deleteS), log.String("exchange", enum.ExchangeTypeName[v]))
			return
		}
	}
}

func (s *compareService) SendNoti() {
	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapExchanges[v]
		if !ok {
			log.Error("compareService GetConnections mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		go func() {
			for message := range ex.GetMsg() {
				s.processMsg(message)
			}
		}()
	}

	go func() {
		for message := range s.priceItemChan {
			s.mergePrice(message)
		}
	}()
}

func (s *compareService) processMsg(msg *ws.MsgChan) {
	switch msg.ExchangeType {
	case enum.ExchangeTypeKucoin:
		s.processKucoinMsg(msg)
	// TODO
	// case enum.ExchangeTypeKucoinFuture:
	// 	s.processKucoinFutureMsg(msg)
	case enum.ExchangeTypeMexc:
		s.processMexcSpotMsg(msg)
	case enum.ExchangeTypeMexcFuture:
		s.processMexcFutureMsg(msg)
	case enum.ExchangeTypeBinanceFuture, enum.ExchangeTypeBinance:
		s.processBinanceMsg(msg)
	case enum.ExchangeTypeOkx:
		s.processOkxMsg(msg)
	default:
		log.Warn("compareService processMsg missing handler", log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]), log.ByteString("msg", msg.Msg))
	}
}

func (s *compareService) processKucoinMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *kdto.MarketTicker

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processKucoinMsg Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Topic, constants.KucoinTopicMarketTicker, "")

	price, err := tickerMsg.Data.Price.Float64()
	if err != nil {
		log.Error("processKucoinMsg parse price error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		Symbol:       symbol,
		Price:        price,
		ExchangeType: msg.ExchangeType,
		At:           time.UnixMilli(tickerMsg.Data.Time),
		TradingType:  msg.TradingType,
	}
}

// func (s *compareService) processKucoinFutureMsg(msg *ws.MsgChan) {
// 	message := msg.Msg

// 	var tickerMsg *kdto.FutureMarketLv2Message

// 	err := json.Unmarshal(message, &tickerMsg)
// 	if err != nil {
// 		log.Error("processKucoinMsg Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
// 		return
// 	}

// 	symbol := strings.ReplaceAll(tickerMsg.Topic, constants.KucoinFutureTopicLv2Market, "")
// 	symbol = strings.ReplaceAll(symbol, constants.CoinUSDT, constants.CoinSymbolSeperateChar+constants.CoinUSDT)

// 	// tickerMsg.Data.Change

// 	// s.priceItemChan <- &dto.ComparePriceChanMsg{
// 	// 	Symbol:       symbol,
// 	// 	Price:        price,
// 	// 	ExchangeType: msg.ExchangeType,
// 	// 	At:           time.UnixMilli(tickerMsg.Data.Time),
// 	// 	TradingType:  msg.TradingType,
// 	// }
// }

func (s *compareService) processMexcFutureMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *mdto.TickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processMexcFutureMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Symbol, "_", constants.CoinSymbolSeperateChar)

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        tickerMsg.Data.LastPrice,
		At:           time.UnixMilli(tickerMsg.Data.Timestamp),
	}
}

func (s *compareService) processMexcSpotMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *mdto.WsSpotBookTickerMsg

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processMexcSpotMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.S, constants.CoinUSDT, constants.CoinSymbolSeperateChar+constants.CoinUSDT)

	price, err := tickerMsg.Data.BuyPrice.Float64()
	if err != nil {
		log.Error("processMexcSpotMsg parse price error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.T),
	}
}

func (s *compareService) processBinanceMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *bdto.TickerMsg

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processBinanceMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Symbol, constants.CoinUSDT, constants.CoinSymbolSeperateChar+constants.CoinUSDT)

	price, err := tickerMsg.LastPrice.Float64()
	if err != nil {
		log.Error("processBinanceMsg parse price error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.CloseTime),
	}
}

func (s *compareService) processOkxMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *okxdto.TickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processOkxMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	if len(tickerMsg.Data) == 0 {
		return
	}

	item := tickerMsg.Data[0]

	price, err := item.IdxPx.Float64()
	if err != nil {
		log.Error("processOkxMsg parse price error", log.ByteString("msg", message))
		return
	}

	ts, err := item.Ts.Int64()
	if err != nil {
		log.Error("processOkxMsg parse ts error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       item.InstId,
		Price:        price,
		At:           time.UnixMilli(ts),
	}
}

func (s *compareService) mapPrice(item *dto.CompareSymbolNotiItem, msg *dto.ComparePriceChanMsg) {
	mapItems := map[enum.ExchangeType]*dto.CompareSymbolNotiExchangeItem{}

	list := item.SpotPrice
	if msg.TradingType == enum.TradingTypeFuture {
		list = item.FuturePrices
	}

	for _, v := range list {
		mapItems[v.ExchangeType] = v
	}

	currentPrice, ok := mapItems[msg.ExchangeType]
	if !ok {
		newPrice := &dto.CompareSymbolNotiExchangeItem{
			ExchangeType: msg.ExchangeType,
			Price:        msg.Price,
			LastPriceAt:  msg.At,
			LastNotiAt:   time.Time{},
		}
		list = append(list, newPrice)

		if msg.TradingType == enum.TradingTypeFuture {
			item.FuturePrices = list
		} else {
			item.SpotPrice = list
		}

		return
	}

	currentPrice.Price = msg.Price
	currentPrice.LastPriceAt = msg.At
}

func (s *compareService) validNoti(item *dto.CompareSymbolNotiItem) {
	if len(item.SpotPrice) == 0 || len(item.FuturePrices) == 0 {
		return
	}

	for _, sp := range item.SpotPrice {
		if time.Since(sp.LastNotiAt) <= constants.MinIntervalNoti {
			continue
		}

		fList := []*dto.CompareSymbolNotiExchangeItem{}

		for _, f := range item.FuturePrices {
			if sp.LastPriceAt.Sub(f.LastPriceAt).Abs().Seconds() > float64(constants.MaxPriceTimeDiffSeconds) {
				continue
			}

			percent := helper.PercentageChange(sp.Price, f.Price)
			otPercent := helper.PercentageChange(f.Price, sp.Price)
			invalidP := math.Abs(percent) < constants.MinDiffPercent
			invalidOt := math.Abs(otPercent) < constants.MinDiffPercent
			if invalidP && invalidOt {
				continue
			}

			if !invalidOt {
				f.Percent = percent
			} else {
				f.Percent = otPercent
			}

			fList = append(fList, f)
		}

		if len(fList) == 0 {
			continue
		}

		sp.LastNotiAt = time.Now()

		err := retry.Do(
			func() error {
				bot, err := tgbotapi.NewBotAPI(s.configs.Telegram.BotAPIKey)
				if err != nil {
					return err
				}

				bot.Debug = true

				u := tgbotapi.NewUpdate(0)
				u.Timeout = 60

				tz, _ := time.LoadLocation(constants.TimeZoneHCM)

				text := fmt.Sprintf(`
					%s - Ex: %s - Type: %s
					Price: %.6f
					Time: %s,
				`, item.Symbol, enum.ExchangeTypeName[sp.ExchangeType], enum.TradingTypeName[enum.TradingTypeSpot],
					sp.Price,
					sp.LastPriceAt.In(tz).Format(constants.TimeHHMMSSFormat),
				)

				for _, v := range fList {
					text += fmt.Sprintf("\n - %s : %s - P: %.6f - T :%s",
						enum.ExchangeTypeName[v.ExchangeType], fmt.Sprintf("%.2f", v.Percent)+"%", v.Price,
						v.LastPriceAt.In(tz).Format(constants.TimeHHMMSSFormat))
				}

				msg := tgbotapi.NewMessage(s.configs.Telegram.ChatID, text)
				_, err = bot.Send(msg)

				return err
			},
			retry.Delay(time.Second),
			retry.LastErrorOnly(true),
		)

		if err != nil {
			log.Error("processMsg tgbotapi error", log.Any("error", err))
			return
		}
	}

}

func (s *compareService) mergePrice(msg *dto.ComparePriceChanMsg) {
	priceItem, ok := s.mapSymbolItem[msg.Symbol]
	if !ok {
		log.Error("processMsg mapSymbolItem not found", log.String("symbol", msg.Symbol), log.Any("item", msg))
		return
	}

	s.mapPrice(priceItem, msg)

	s.validNoti(priceItem)
}
