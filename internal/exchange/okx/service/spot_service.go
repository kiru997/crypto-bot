package service

import (
	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/ws"
)

type SpotService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.Msg
	GetConnections() map[string]*idto.ConnectionItem
}

func NewSpotService(configs *configs.AppConfig) SpotService {
	return ws.NewWS(NewSpotExchange(configs))
}

// func processOkxMsg(msg *ws.Msg) {
// 	message := msg.Msg

// 	var tickerMsg *dto.TickerMessage

// 	err := json.Unmarshal(message, &tickerMsg)
// 	if err != nil {
// 		log.Error("processOkxMsg error Unmarshal error", log.Any("error", err), log.ByteString("msg", message))
// 		return
// 	}

// 	if len(tickerMsg.Data) == 0 {
// 		return
// 	}

// 	item := tickerMsg.Data[0]

// 	price, err := item.IdxPx.Float64()
// 	if err != nil {
// 		log.Error("processOkxMsg parse price error", log.ByteString("msg", message))
// 		return
// 	}

// 	ts, err := item.Ts.Int64()
// 	if err != nil {
// 		log.Error("processOkxMsg parse ts error", log.ByteString("msg", message))
// 		return
// 	}

// 	s.priceItemChan <- &idto.ComparePriceChanMsg{
// 		ExchangeType: msg.ExchangeType,
// 		TradingType:  msg.TradingType,
// 		Symbol:       item.InstId,
// 		Price:        price,
// 		At:           time.UnixMilli(ts),
// 		ConID:        msg.ConnID,
// 	}
// }
