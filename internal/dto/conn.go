package dto

import (
	"time"

	"example.com/greetings/pkg/enum"
	"example.com/greetings/pkg/log"
	"github.com/gorilla/websocket"
)

type ConnectionItem struct {
	T            time.Time
	Conn         *websocket.Conn
	Symbols      []string
	Ticker       *time.Ticker
	ExchangeType enum.ExchangeType
	TradingType  enum.TradingType
}

func (c *ConnectionItem) Close() {
	c.Ticker.Stop()
	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Error("ConnectionItem WriteMessage close error", log.Any("error", err))
	}
	c.Conn.Close()
}
