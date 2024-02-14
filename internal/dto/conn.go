package dto

import (
	"time"

	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"github.com/gorilla/websocket"
)

type ConnectionItem struct {
	ID           string
	T            time.Time
	Conn         *websocket.Conn
	Symbols      []string
	Ticker       *time.Ticker
	ExchangeType enum.ExchangeType
	TradingType  enum.TradingType
	Done         bool
}

func (c *ConnectionItem) Close() {
	c.Done = true
	c.Ticker.Stop()

	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Error("ConnectionItem Close CloseMessage error", log.Any("error", err), log.String("id", c.ID))
	}

	if err = c.Conn.Close(); err != nil {
		log.Error("ConnectionItem Close close error", log.Any("error", err), log.String("id", c.ID))
	}
}
