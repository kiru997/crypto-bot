package service

import (
	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/ws"
)

type FutureService interface {
	Subcribe(symbols []string) error
	UnSubcribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.MsgChan
	GetConnections() map[string]*idto.ConnectionItem
}

func NewFutureService(configs *configs.AppConfig) FutureService {
	return ws.NewWS(NewExchange(configs))
}
