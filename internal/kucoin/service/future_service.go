package service

import (
	idto "example.com/greetings/internal/dto"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/ws"
	kumex "github.com/Kucoin/kucoin-futures-go-sdk"
)

type KucoinFutureService interface {
	Subscribe(symbols []string) error
	UnSubscribe(symbols []string) error
	RefreshConn()
	GetMsg() chan *ws.MsgChan
	GetConnections() map[string]*idto.ConnectionItem
}

type kucoinFutureService struct {
	configs  *configs.AppConfig
	c        *kumex.ApiService
	exchange ws.WS
}

func NewKucoinFutureService(configs *configs.AppConfig, c *kumex.ApiService) KucoinFutureService {
	return &kucoinFutureService{
		configs:  configs,
		c:        c,
		exchange: ws.NewWS(NewKucoinFutureExchange(configs, c)),
	}
}

func (s *kucoinFutureService) ActiveContracts() (kumex.ContractsModels, error) {
	resp, err := s.c.ActiveContracts()
	if err != nil {
		return nil, err
	}

	os := &kumex.ContractsModels{}
	resp.ReadData(&os)

	return *os, nil
}

func (s *kucoinFutureService) GetConnections() map[string]*idto.ConnectionItem {
	return s.exchange.GetConnections()
}

func (s *kucoinFutureService) GetMsg() chan *ws.MsgChan {
	return s.exchange.GetMsg()
}

func (s *kucoinFutureService) RefreshConn() {
	s.exchange.RefreshConn()
}

func (s *kucoinFutureService) Subscribe(symbols []string) error {
	return s.exchange.Subscribe(symbols)
}

func (s *kucoinFutureService) UnSubscribe(symbols []string) error {
	return s.exchange.UnSubscribe(symbols)
}
