package dto

import "encoding/json"

type TickerMsg struct {
	EventType          string      `json:"e"`
	EventTime          int64       `json:"E"`
	Symbol             string      `json:"s"`
	PriceChange        string      `json:"p"`
	PriceChangePercent string      `json:"P"`
	WeightedAvgPrice   string      `json:"w"`
	LastPrice          json.Number `json:"c"`
	CloseQty           string      `json:"Q"`
	OpenPrice          string      `json:"o"`
	HighPrice          string      `json:"h"`
	LowPrice           string      `json:"l"`
	Volume             string      `json:"v"`
	QuoteVolume        string      `json:"q"`
	OpenTime           int64       `json:"O"`
	CloseTime          int64       `json:"C"`
	FirstTradeID       int         `json:"F"`
	LastTradeID        int         `json:"L"`
	TradeCount         int         `json:"n"`
}
