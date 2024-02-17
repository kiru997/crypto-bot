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
	QuoteVolume        json.Number `json:"q"`
	OpenTime           int64       `json:"O"`
	CloseTime          int64       `json:"C"`
	FirstTradeID       int         `json:"F"`
	LastTradeID        int         `json:"L"`
	TradeCount         int         `json:"n"`
}

type SpotTicker24hRes []*SpotTicker24h

type SpotTicker24h struct {
	Symbol             string      `json:"symbol"`
	PriceChange        string      `json:"priceChange"`
	PriceChangePercent json.Number `json:"priceChangePercent"`
	WeightedAvgPrice   string      `json:"weightedAvgPrice"`
	PrevClosePrice     string      `json:"prevClosePrice"`
	LastPrice          string      `json:"lastPrice"`
	LastQty            string      `json:"lastQty"`
	BidPrice           string      `json:"bidPrice"`
	BidQty             string      `json:"bidQty"`
	AskPrice           string      `json:"askPrice"`
	AskQty             string      `json:"askQty"`
	OpenPrice          string      `json:"openPrice"`
	HighPrice          string      `json:"highPrice"`
	LowPrice           string      `json:"lowPrice"`
	Volume             string      `json:"volume"`
	QuoteVolume        json.Number `json:"quoteVolume"`
	OpenTime           int64       `json:"openTime"`
	CloseTime          int64       `json:"closeTime"`
	FirstID            int64       `json:"firstId"`
	LastID             int64       `json:"lastId"`
	Count              int64       `json:"count"`
}
