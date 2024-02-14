package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	bdto "example.com/greetings/internal/binance/dto"
	bydto "example.com/greetings/internal/bybit/dto"
	"example.com/greetings/internal/dto"
	gdto "example.com/greetings/internal/gate/dto"
	kdto "example.com/greetings/internal/kucoin/dto"
	mdto "example.com/greetings/internal/mexc/dto"
	okxdto "example.com/greetings/internal/okx/dto"

	bs "example.com/greetings/internal/binance/service"
	bys "example.com/greetings/internal/bybit/service"
	gs "example.com/greetings/internal/gate/service"
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
	WatchTopChange(string)
	SendNoti()
	RefreshConn()
	GetCurrentSymbolItems() []*dto.CompareSymbolNotiItem
	GetConnections() map[string]*dto.ConnectionItem
}

type ExchangeService interface {
	TopChange(context.Context) ([]string, error)
}

type compareService struct {
	configs            *configs.AppConfig
	compareConfigs     *configs.CompareConfig
	mapSymbolItem      map[string]*dto.CompareSymbolNotiItem
	priceItemChan      chan *dto.ComparePriceChanMsg
	mapExchanges       map[enum.ExchangeType]ws.WS
	mapExchangeService map[enum.ExchangeType]ExchangeService
	mapExchangeCompare map[enum.ExchangeType][]enum.ExchangeType
}

func NewCompareService(configs *configs.AppConfig,
	compareConfigs *configs.CompareConfig,
	kucoinSpotService ks.KucoinSpotService,
	kucoinFutureService ks.KucoinFutureService,
	binanceService bs.SpotService,
	binanceFutureService bs.FutureService,
	mexcFutureService ms.FutureService,
	mexcSpotService ms.SpotService,
	okxFutureService okxs.FutureService,
	bybitService bys.SpotService,
	bybitFutureService bys.FutureService,
	gateService gs.SpotService,
	gateFutureService gs.FutureService,
) CompareService {

	res := map[enum.ExchangeType][]enum.ExchangeType{}
	for _, v := range *compareConfigs {
		if !v.Enable {
			continue
		}
		res[v.Exchange] = append(res[v.Exchange], v.FutureExchanges...)
	}

	return &compareService{
		configs:        configs,
		compareConfigs: compareConfigs,
		mapSymbolItem:  map[string]*dto.CompareSymbolNotiItem{},
		priceItemChan:  make(chan *dto.ComparePriceChanMsg),
		mapExchanges: map[enum.ExchangeType]ws.WS{
			enum.ExchangeTypeOkx:           okxFutureService,
			enum.ExchangeTypeKucoin:        kucoinSpotService,
			enum.ExchangeTypeKucoinFuture:  kucoinFutureService,
			enum.ExchangeTypeMexc:          mexcSpotService,
			enum.ExchangeTypeMexcFuture:    mexcFutureService,
			enum.ExchangeTypeBinance:       binanceService,
			enum.ExchangeTypeBinanceFuture: binanceFutureService,
			enum.ExchangeTypeBybit:         bybitService,
			enum.ExchangeTypeBybitFuture:   bybitFutureService,
			enum.ExchangeTypeGate:          gateService,
			enum.ExchangeTypeGateFuture:    gateFutureService,
		},
		mapExchangeService: map[enum.ExchangeType]ExchangeService{
			enum.ExchangeTypeKucoin:  kucoinSpotService,
			enum.ExchangeTypeMexc:    mexcSpotService,
			enum.ExchangeTypeBinance: binanceService,
			enum.ExchangeTypeBybit:   bybitService,
			enum.ExchangeTypeGate:    gateService,
		},
		mapExchangeCompare: res,
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

func (s *compareService) WatchTopChange(jobID string) {
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

	symbolSubscribes := make([]string, 0, len(s.mapSymbolItem))

	for _, v := range s.mapSymbolItem {
		symbolSubscribes = append(symbolSubscribes, v.Symbol)
	}

	add, deleteS := lo.Difference(symbols, symbolSubscribes)

	for _, v := range add {
		s.mapSymbolItem[v] = &dto.CompareSymbolNotiItem{
			Symbol: v,
		}
	}

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapExchanges[v]
		if !ok {
			log.Error("compareService TopChange mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		err := ex.UnSubscribe(deleteS)
		if err != nil {
			log.Error("TopChange UnSubscribe error", log.Any("error", err), log.Any("symbols", deleteS), log.String("exchange", enum.ExchangeTypeName[v]))
			return
		}

		err = ex.Subscribe(add)
		if err != nil {
			log.Error("TopChange Subscribe error", log.Any("error", err), log.Any("symbols", add), log.String("exchange", enum.ExchangeTypeName[v]))
			return
		}
	}

	for _, v := range deleteS {
		log.Debug("compareService TopChange remove symbol", log.String("symbol", v), log.String("jID", jobID))
		delete(s.mapSymbolItem, v)
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
	case enum.ExchangeTypeKucoinFuture:
		s.processKucoinFutureMsg(msg)
	case enum.ExchangeTypeMexc:
		s.processMexcSpotMsg(msg)
	case enum.ExchangeTypeMexcFuture:
		s.processMexcFutureMsg(msg)
	case enum.ExchangeTypeBinanceFuture, enum.ExchangeTypeBinance:
		s.processBinanceMsg(msg)
	case enum.ExchangeTypeBybit:
		s.processBybitMsg(msg)
	case enum.ExchangeTypeBybitFuture:
		s.processBybitFutureMsg(msg)
	case enum.ExchangeTypeGate:
		s.processGateSpotMsg(msg)
	case enum.ExchangeTypeGateFuture:
		s.processGateFutureMsg(msg)
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
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processKucoinFutureMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *kdto.WSFutureTickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processKucoinMsg Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Data.Symbol, constants.CoinUSDTM, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

	price, err := tickerMsg.Data.Price.Float64()
	if err != nil {
		log.Error("processKucoinMsg parse price error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		Symbol:       symbol,
		Price:        price,
		ExchangeType: msg.ExchangeType,
		At:           time.UnixMilli(tickerMsg.Data.Timestamp / 1e6),
		TradingType:  msg.TradingType,
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processMexcFutureMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *mdto.TickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processMexcFutureMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Symbol, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar)

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        tickerMsg.Data.LastPrice,
		At:           time.UnixMilli(tickerMsg.Data.Timestamp),
		ConID:        msg.ConnID,
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

	symbol := strings.ReplaceAll(tickerMsg.S, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

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
		ConID:        msg.ConnID,
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

	symbol := strings.ReplaceAll(tickerMsg.Symbol, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

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
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processBybitMsg(msg *ws.MsgChan) {
	message := msg.Msg
	var tickerMsg *bydto.WSSpotTickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processBybitMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message),
			log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Data.Symbol, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

	price, err := tickerMsg.Data.LastPrice.Float64()
	if err != nil {
		log.Error("processBybitMsg parse price error", log.Any("error", err), log.ByteString("msg", message),
			log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.Ts),
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processBybitFutureMsg(msg *ws.MsgChan) {
	message := msg.Msg
	var tickerMsg *bydto.WSFutureTickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processBybitFutureMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message),
			log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Data.Symbol, constants.CoinUSDT, constants.CoinSymbolSeparateChar+constants.CoinUSDT)

	price, err := tickerMsg.Data.Bid1Price.Float64()
	if err != nil {
		log.Error("processBybitMsg parse price error", log.Any("error", err), log.ByteString("msg", message),
			log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.Ts),
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processGateSpotMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *gdto.WSSpotTickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message),
			log.String("cId", msg.ConnID), log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	symbol := strings.ReplaceAll(tickerMsg.Result.CurrencyPair, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar)

	price, err := tickerMsg.Result.Last.Float64()
	if err != nil {
		log.Error("processMsg parse price error", log.Any("error", err), log.ByteString("msg", message),
			log.String("cId", msg.ConnID), log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.TimeMs),
		ConID:        msg.ConnID,
	}
}

func (s *compareService) processGateFutureMsg(msg *ws.MsgChan) {
	message := msg.Msg

	var tickerMsg *gdto.WSFutureTickerMessage

	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		log.Error("processGateFutureMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
		return
	}

	if len(tickerMsg.Result) == 0 {
		log.Warn("processGateFutureMsg empty result", log.ByteString("msg", message))
		return
	}

	first := tickerMsg.Result[0]

	symbol := strings.ReplaceAll(first.Contract, constants.CoinSymbolSeparateCharUnderscore, constants.CoinSymbolSeparateChar)

	price, err := first.Last.Float64()
	if err != nil {
		log.Error("processGateFutureMsg parse price error", log.ByteString("msg", message))
		return
	}

	s.priceItemChan <- &dto.ComparePriceChanMsg{
		ExchangeType: msg.ExchangeType,
		TradingType:  msg.TradingType,
		Symbol:       symbol,
		Price:        price,
		At:           time.UnixMilli(tickerMsg.TimeMs),
		ConID:        msg.ConnID,
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
		ConID:        msg.ConnID,
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
		fList := []*dto.CompareSymbolNotiExchangeItem{}

		exchanges, ok := s.mapExchangeCompare[sp.ExchangeType]
		if !ok {
			continue
		}

		for _, f := range item.FuturePrices {
			_, exist := lo.Find(exchanges, func(ex enum.ExchangeType) bool {
				return ex == f.ExchangeType
			})

			if !exist {
				continue
			}

			if time.Since(f.LastNotiAt) <= constants.MinIntervalNoti {
				continue
			}

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

			f.LastNotiAt = time.Now()
			fList = append(fList, f)
		}

		if len(fList) == 0 {
			continue
		}

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
		log.Error("processMsg mapSymbolItem not found", log.String("symbol", msg.Symbol), log.Any("item", msg),
			log.String("exchange", enum.ExchangeTypeName[msg.ExchangeType]))
		return
	}

	s.mapPrice(priceItem, msg)

	s.validNoti(priceItem)
}
