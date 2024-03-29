package ws

import (
	"sync"
	"time"

	"example.com/greetings/internal/dto"
	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/samber/lo"
)

type ws struct {
	exchange       Exchange
	mapConnections map[string]*dto.ConnectionItem
	mapSymbols     map[string]string
	msgChan        chan *Msg
	l              sync.Mutex
}

type ExChangeConfig struct {
	ExchangeType             enum.ExchangeType
	TradingType              enum.TradingType
	RefreshConnectionMinutes int
	MaxSubscriptions         int
}

type Exchange interface {
	GetConfig() *ExChangeConfig
	GetBaseURL() (string, error)
	GetPingMsg() []byte
	FilterMsg([]byte) bool
	GetSubscribeMsg(string) []byte
	GetUnSubscribeMsg(string) []byte
}

type WS interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *Msg
	GetConnections() map[string]*dto.ConnectionItem
}

type Msg struct {
	ExchangeType enum.ExchangeType
	TradingType  enum.TradingType
	Msg          []byte
	ConnID       string
}

func NewWS(ex Exchange) WS {
	return &ws{
		exchange:       ex,
		mapConnections: make(map[string]*dto.ConnectionItem),
		mapSymbols:     map[string]string{},
		msgChan:        make(chan *Msg),
		l:              sync.Mutex{},
	}
}

func (s *ws) CreateConnection() (*dto.ConnectionItem, error) {
	cID := uuid.NewString()

	url, err := s.exchange.GetBaseURL()
	if err != nil {
		return nil, err
	}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	const pingTime = time.Second * 15

	ticker := time.NewTicker(pingTime)

	go func() {
		for range ticker.C {
			exType := s.exchange.GetConfig().ExchangeType
			if exType == enum.ExchangeTypeBinanceFuture || exType == enum.ExchangeTypeBinance {
				err := c.WriteMessage(websocket.PingMessage, s.exchange.GetPingMsg())
				if err != nil {
					log.Error("ws CreateConnection WriteMessage PingMessage error", log.Any("error", err),
						log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]), log.String("cId", cID))
				}

				continue
			}

			err := c.WriteMessage(websocket.TextMessage, s.exchange.GetPingMsg())
			if err != nil {
				log.Error("ws CreateConnection WriteMessage ping error", log.Any("error", err),
					log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]), log.String("cId", cID))
			}
		}
	}()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				conn, existConn := s.mapConnections[cID]
				if !existConn {
					return
				}

				if conn.Done {
					return
				}

				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					connIDs := []string{}

					s.l.Lock()

					for k := range s.mapConnections {
						connIDs = append(connIDs, k)
					}

					s.l.Unlock()
					log.Error("ws CreateConnection ReadMessage error", log.String("cId", cID), log.Any("error", err), log.Any("conn", connIDs),
						log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
				}

				break
			}

			if s.exchange.FilterMsg(message) {
				continue
			}

			m := &Msg{
				ExchangeType: s.exchange.GetConfig().ExchangeType,
				TradingType:  s.exchange.GetConfig().TradingType,
				Msg:          message,
				ConnID:       cID,
			}

			s.msgChan <- m
		}
	}()

	return &dto.ConnectionItem{
		ID:           cID,
		T:            time.Now(),
		Conn:         c,
		Symbols:      []string{},
		Ticker:       ticker,
		ExchangeType: s.exchange.GetConfig().ExchangeType,
		TradingType:  s.exchange.GetConfig().TradingType,
	}, nil
}

// func keepAlive(c *websocket.Conn, timeout time.Duration) {
// 	ticker := time.NewTicker(timeout)

// 	lastResponse := time.Now()
// 	c.SetPongHandler(func(msg string) error {
// 		lastResponse = time.Now()
// 		return nil
// 	})

// 	go func() {
// 		defer ticker.Stop()
// 		for {
// 			deadline := time.Now().Add(10 * time.Second)
// 			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
// 			if err != nil {
// 				return
// 			}
// 			<-ticker.C
// 			if time.Since(lastResponse) > timeout {
// 				c.Close()
// 				return
// 			}
// 		}
// 	}()
// }

func (s *ws) RefreshConn() {
	symbols := []string{}

	for id, v := range s.mapConnections {
		if time.Since(v.T) < time.Duration(s.exchange.GetConfig().RefreshConnectionMinutes)*time.Minute {
			continue
		}

		symbols = append(symbols, v.Symbols...)

		_ = s.UnSubscribe(v.Symbols)
		v.Close()

		delete(s.mapConnections, id)

		log.Debug("ws RefreshConn delete conn", log.String("id", id), log.Time("initAt", v.T),
			log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}

	err := s.Subscribe(symbols)
	if err != nil {
		log.Error("ws RefreshConn error", log.Any("error", err), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}
}

func (s *ws) Subscribe(symbols []string) error {
	dupSym := lo.FindDuplicates(symbols)
	log.Debug("Subscribe items", log.Any("symbols", symbols), log.Any("dup", dupSym))

	if len(symbols) == 0 {
		return nil
	}

	s.l.Lock()

	defer s.l.Unlock()

	for _, symbol := range symbols {
		skipInit := false
		msg := s.exchange.GetSubscribeMsg(symbol)

		for connID, conn := range s.mapConnections {
			if len(conn.Symbols) >= s.exchange.GetConfig().MaxSubscriptions {
				continue
			}

			err := conn.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Error("ws Subscribe WriteMessage subscribe symbol error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			} else {
				conn.Symbols = append(conn.Symbols, symbol)
				s.mapSymbols[symbol] = connID
			}

			log.Debug("ws Subscribe done", log.String("symbol", symbol), log.String("id", connID),
				log.Int("numberConnections", len(s.mapConnections)), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))

			skipInit = true

			break
		}

		if skipInit {
			continue
		}

		connItem, err := s.CreateConnection()
		if err != nil {
			log.Error("ws Subscribe CreateConnection error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		err = connItem.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error("ws Subscribe WriteMessage ping error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		} else {
			connItem.Symbols = append(connItem.Symbols, symbol)
			s.mapSymbols[symbol] = connItem.ID
		}

		s.mapConnections[connItem.ID] = connItem
		log.Debug("ws Subscribe done", log.String("id", connItem.ID),
			log.String("symbol", symbol), log.Int("numberConnections", len(s.mapConnections)), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}

	return nil
}

func (s *ws) UnSubscribe(symbols []string) error {
	log.Debug("UnSubscribe items", log.Any("symbols", symbols))
	s.l.Lock()

	defer s.l.Unlock()

	for _, symbol := range symbols {
		connID, ok := s.mapSymbols[symbol]
		if !ok {
			log.Warn("ws UnSubscribe not found map Symbol", log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		conn, ok := s.mapConnections[connID]
		if !ok {
			log.Warn("ws UnSubscribe not found map connections",
				log.String("connId", connID), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		err := conn.Conn.WriteMessage(websocket.TextMessage, s.exchange.GetUnSubscribeMsg(symbol))
		if err != nil {
			log.Error("ws UnSubscribe WriteMessage error", log.Any("error", err), log.String("symbol", symbol),
				log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		}

		conn.Symbols = lo.Filter(conn.Symbols, func(sym string, _ int) bool {
			return sym != symbol
		})

		delete(s.mapSymbols, symbol)

		log.Debug("ws UnSubscribe done", log.String("connId", connID),
			log.String("symbol", symbol), log.Int("numberConnections", len(s.mapConnections)),
			log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}

	return nil
}

func (s *ws) GetMsg() chan *Msg {
	return s.msgChan
}

func (s *ws) GetConnections() map[string]*dto.ConnectionItem {
	return s.mapConnections
}
