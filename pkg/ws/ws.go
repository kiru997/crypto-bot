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
	msgChan        chan *MsgChan
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
	GetSubcribeMsg(string) []byte
	GetUnSubcribeMsg(string) []byte
}

type WS interface {
	Subcribe(symbols []string) error
	UnSubcribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *MsgChan
	GetConnections() map[string]*dto.ConnectionItem
}

type MsgChan struct {
	ExchangeType enum.ExchangeType
	TradingType  enum.TradingType
	Msg          []byte
}

func NewWS(ex Exchange) WS {
	return &ws{
		exchange:       ex,
		mapConnections: make(map[string]*dto.ConnectionItem),
		mapSymbols:     map[string]string{},
		msgChan:        make(chan *MsgChan),
		l:              sync.Mutex{},
	}
}

func (s *ws) CreateConnection() (string, *websocket.Conn, *time.Ticker, error) {
	cID := uuid.NewString()

	url, err := s.exchange.GetBaseURL()
	if err != nil {
		return "", nil, nil, err
	}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return "", nil, nil, err
	}

	const pingTime = time.Second * 15

	ticker := time.NewTicker(pingTime)

	go func() {
		for range ticker.C {
			exType := s.exchange.GetConfig().ExchangeType
			if exType == enum.ExchangeTypeBinanceFuture || exType == enum.ExchangeTypeBinance {
				// const deadline = 10 * time.Second
				// err := c.WriteControl(websocket.PingMessage, s.exchange.GetPingMsg(), time.Now().Add(deadline))
				// if err != nil {
				// 	log.Error("ws CreateConnection WriteControl ping error", log.Any("error", err),
				// 		log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]), log.String("cId", cID))
				// }

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
				connIDs := []string{}

				s.l.Lock()

				for k := range s.mapConnections {
					connIDs = append(connIDs, k)
				}

				s.l.Unlock()
				log.Error("ws CreateConnection ReadMessage error", log.String("cId", cID), log.Any("error", err), log.Any("conn", connIDs),
					log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))

				return
			}

			if s.exchange.FilterMsg(message) {
				continue
			}

			s.msgChan <- &MsgChan{
				ExchangeType: s.exchange.GetConfig().ExchangeType,
				TradingType:  s.exchange.GetConfig().TradingType,
				Msg:          message,
			}

		}
	}()

	return cID, c, ticker, nil
}

func (s *ws) RefreshConn() {
	s.l.Lock()

	defer s.l.Unlock()

	for id, v := range s.mapConnections {
		if time.Since(v.T) < time.Duration(s.exchange.GetConfig().RefreshConnectionMinutes)*time.Minute {
			continue
		}

		log.Debug("ws RefreshConn run", log.String("id", id), log.Time("initAt", v.T),
			log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		v.Close()

		symbols := v.Symbols

		delete(s.mapConnections, id)

		err := s.Subcribe(symbols)
		if err != nil {
			log.Error("ws RefreshConn error", log.Any("error", err), log.String("id", id), log.Time("initAt", v.T),
				log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		}
	}
}

func (s *ws) Subcribe(symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	for _, symbol := range symbols {
		skipInit := false
		msg := s.exchange.GetSubcribeMsg(symbol)

		for connID, conn := range s.mapConnections {
			if len(conn.Symbols) >= s.exchange.GetConfig().MaxSubscriptions {
				continue
			}

			err := conn.Conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Error("ws Subcribe WriteMessage subcribe symbol error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			} else {
				conn.Symbols = append(conn.Symbols, symbol)
				s.mapSymbols[symbol] = connID
			}
			log.Debug("ws Subcribe done", log.String("symbol", symbol), log.Int("numberConnections", len(s.mapConnections)), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))

			skipInit = true
		}

		if skipInit {
			continue
		}

		connID, c, ticker, err := s.CreateConnection()
		if err != nil {
			log.Error("ws Subcribe CreateConnection error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		cI := &dto.ConnectionItem{
			T:            time.Now(),
			Conn:         c,
			Ticker:       ticker,
			ExchangeType: s.exchange.GetConfig().ExchangeType,
			TradingType:  s.exchange.GetConfig().TradingType,
		}

		err = c.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Error("ws Subcribe WriteMessage ping error", log.Any("error", err), log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		} else {
			cI.Symbols = append(cI.Symbols, symbol)
			s.mapSymbols[symbol] = connID
		}

		s.mapConnections[connID] = cI

		log.Debug("ws Subcribe done", log.String("symbol", symbol), log.Int("numberConnections", len(s.mapConnections)), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}

	return nil
}

func (s *ws) UnSubcribe(symbols []string) error {
	for _, symbol := range symbols {
		connID, ok := s.mapSymbols[symbol]
		if !ok {
			log.Warn("ws UnSubcribe not found map Symbol", log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		conn, ok := s.mapConnections[connID]
		if !ok {
			log.Warn("ws UnSubcribe not found map connections", log.String("symbol", symbol), log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
			continue
		}

		err := conn.Conn.WriteMessage(websocket.TextMessage, s.exchange.GetUnSubcribeMsg(symbol))
		if err != nil {
			log.Error("ws UnSubcribe WriteMessage error", log.Any("error", err), log.String("symbol", symbol),
				log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
		}

		conn.Symbols = lo.Filter(conn.Symbols, func(sym string, _ int) bool {
			return sym != symbol
		})

		delete(s.mapSymbols, symbol)

		if len(conn.Symbols) == 0 {
			conn.Close()
			delete(s.mapConnections, connID)
		}

		log.Debug("ws UnSubcribe done", log.String("newConnId", connID),
			log.String("symbol", symbol), log.Int("numberConnections", len(s.mapConnections)),
			log.String("exchange", enum.ExchangeTypeName[s.exchange.GetConfig().ExchangeType]))
	}

	initSymbol := []string{}
	for cID, v := range s.mapConnections {
		initSymbol = append(initSymbol, v.Symbols...)
		v.Close()
		delete(s.mapConnections, cID)
	}

	return s.Subcribe(initSymbol)
}

func (s *ws) GetMsg() chan *MsgChan {
	return s.msgChan
}

func (s *ws) GetConnections() map[string]*dto.ConnectionItem {
	return s.mapConnections
}
