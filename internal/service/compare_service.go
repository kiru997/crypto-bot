package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"example.com/greetings/internal/dto"
	"golang.org/x/sync/errgroup"

	bs "example.com/greetings/internal/exchange/binance/service"
	bms "example.com/greetings/internal/exchange/bitmart/service"
	bys "example.com/greetings/internal/exchange/bybit/service"
	gs "example.com/greetings/internal/exchange/gate/service"
	ks "example.com/greetings/internal/exchange/kucoin/service"
	ms "example.com/greetings/internal/exchange/mexc/service"
	okxs "example.com/greetings/internal/exchange/okx/service"

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

type BaseCompareService interface {
	TopChange(context.Context) ([]string, error)
}

type WSTickerService interface {
	ws.WS
	ProcessTickerMsg(cha chan *dto.ComparePriceChanMsg)
}

type compareService struct {
	configs            *configs.AppConfig
	compareConfigs     *configs.CompareConfig
	mapSymbolItem      map[string]*dto.CompareSymbolNotiItem
	priceItemChan      chan *dto.ComparePriceChanMsg
	mapExchangeCompare map[enum.ExchangeType][]enum.ExchangeType
	mapTickerService   map[enum.ExchangeType]WSTickerService
}

func NewCompareService(configs *configs.AppConfig,
	compareConfigs *configs.CompareConfig,
	kucoinSpotService ks.SpotService,
	kucoinFutureService ks.FutureService,
	binanceService bs.SpotService,
	binanceFutureService bs.FutureService,
	mexcFutureService ms.FutureService,
	mexcSpotService ms.SpotService,
	bybitService bys.SpotService,
	bybitFutureService bys.FutureService,
	gateService gs.SpotService,
	gateFutureService gs.FutureService,
	bitmartService bms.SpotService,
	bitmartFutureService bms.FutureService,
	okxService okxs.SpotService,
) CompareService {
	res := map[enum.ExchangeType][]enum.ExchangeType{}

	for _, v := range *compareConfigs {
		if !v.Enable {
			continue
		}

		res[v.Exchange] = append(res[v.Exchange], v.CompareExchanges...)
	}

	return &compareService{
		configs:        configs,
		compareConfigs: compareConfigs,
		mapSymbolItem:  map[string]*dto.CompareSymbolNotiItem{},
		priceItemChan:  make(chan *dto.ComparePriceChanMsg),
		mapTickerService: map[enum.ExchangeType]WSTickerService{
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
			enum.ExchangeTypeBitmart:       bitmartService,
			enum.ExchangeTypeBitmartFuture: bitmartFutureService,
			enum.ExchangeTypeOkx:           okxService,
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
		res = append(res, v.CompareExchanges...)
	}

	res = lo.Uniq(res)

	return res
}

func (s *compareService) RefreshConn() {
	log.Info("RefreshConn start")

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapTickerService[v]
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
		ex, ok := s.mapTickerService[v]
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

	jCtx, cancel := context.WithTimeout(context.Background(), time.Minute)

	defer cancel()

	g, ctx := errgroup.WithContext(jCtx)

	for _, v := range s.GetBaseCompareExchange() {
		ex, ok := s.mapTickerService[v]
		if !ok {
			log.Error("compareService WatchTopChange mapExchangeService not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		g.Go(func() error {
			bcs, ok := ex.(BaseCompareService)
			if !ok {
				log.Error("compareService convert BaseCompareService TopChange error", log.String("exchange", enum.ExchangeTypeName[v]))
				return nil
			}

			s, err := bcs.TopChange(ctx)
			if err != nil {
				log.Error("compareService WatchTopChange TopChange error", log.Any("error", err), log.String("exchange", enum.ExchangeTypeName[v]))
				return nil
			}

			symbols = append(symbols, s...)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		log.Error("compareService WatchTopChange Wait TopChange error", log.Any("error", err))
		return
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

	var wg sync.WaitGroup

	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapTickerService[v]
		if !ok {
			log.Error("compareService TopChange mapExchange not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		err := ex.UnSubscribe(deleteS)
		if err != nil {
			log.Error("TopChange UnSubscribe error", log.Any("error", err), log.Any("symbols", deleteS), log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			err = ex.Subscribe(add)
			if err != nil {
				log.Error("TopChange Subscribe error", log.Any("error", err), log.Any("symbols", add), log.String("exchange", enum.ExchangeTypeName[v]))
				return
			}
		}()
	}

	wg.Wait()

	for _, v := range deleteS {
		log.Debug("compareService TopChange remove symbol", log.String("symbol", v), log.String("jID", jobID))
		delete(s.mapSymbolItem, v)
	}
}

func (s *compareService) SendNoti() {
	go s.ProcessMsg()
	go func() {
		for message := range s.priceItemChan {
			s.mergePrice(message)
		}
	}()
}

func (s *compareService) ProcessMsg() {
	for _, v := range s.GetAvailExchange() {
		ex, ok := s.mapTickerService[v]
		if !ok {
			log.Error("compareService processMsg mapWSHandler not found error", log.String("exchange", enum.ExchangeTypeName[v]))
			continue
		}

		go ex.ProcessTickerMsg(s.priceItemChan)
	}
}

func (s *compareService) mapPrice(item *dto.CompareSymbolNotiItem, msg *dto.ComparePriceChanMsg) {
	mapItems := map[enum.ExchangeType]*dto.CompareSymbolNotiExchangeItem{}

	list := item.SpotPrices
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
			item.SpotPrices = list
		}

		return
	}

	currentPrice.Price = msg.Price
	currentPrice.LastPriceAt = msg.At
}

func (s *compareService) notify(symbol string, baseItem *dto.CompareSymbolNotiExchangeItem, fList []*dto.CompareSymbolNotiExchangeItem) {
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
				%s - Ex: %s
				Price: %.6f
				Time: %s,
			`, symbol, enum.ExchangeTypeName[baseItem.ExchangeType],
				baseItem.Price,
				baseItem.LastPriceAt.In(tz).Format(constants.TimeHHMMSSFormat),
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

func (s *compareService) compare(symbol string, from, to []*dto.CompareSymbolNotiExchangeItem) {
	for _, sp := range from {
		fList := []*dto.CompareSymbolNotiExchangeItem{}

		exchanges, ok := s.mapExchangeCompare[sp.ExchangeType]
		if !ok {
			continue
		}

		for _, f := range to {
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

		s.notify(symbol, sp, fList)
	}
}

func (s *compareService) validNoti(item *dto.CompareSymbolNotiItem) {
	if len(item.SpotPrices) == 0 || len(item.FuturePrices) == 0 {
		return
	}

	s.compare(item.Symbol, item.SpotPrices, item.FuturePrices)
	s.compare(item.Symbol, item.FuturePrices, item.SpotPrices)

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
