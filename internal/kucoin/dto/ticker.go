package dto

import "encoding/json"

type MarketTicker struct {
	Topic   string     `json:"topic"`
	Type    string     `json:"type"`
	Data    TickerData `json:"data"`
	Subject string     `json:"subject"`
}

type TickerData struct {
	BestAsk     string      `json:"bestAsk"`
	BestAskSize string      `json:"bestAskSize"`
	BestBid     string      `json:"bestBid"`
	BestBidSize string      `json:"bestBidSize"`
	Price       json.Number `json:"price"`
	Sequence    string      `json:"sequence"`
	Size        string      `json:"size"`
	Time        int64       `json:"time"`
}

type WSFutureTickerData struct {
	Symbol       string      `json:"symbol"`
	Sequence     int64       `json:"sequence"`
	Side         string      `json:"side"`
	Size         int         `json:"size"`
	Price        json.Number `json:"price"`
	BestBidSize  int         `json:"bestBidSize"`
	BestBidPrice string      `json:"bestBidPrice"`
	BestAskPrice string      `json:"bestAskPrice"`
	TradeID      string      `json:"tradeId"`
	BestAskSize  int         `json:"bestAskSize"`
	Timestamp    int64       `json:"ts"`
}

type WSFutureTickerMessage struct {
	Topic   string             `json:"topic"`
	Type    string             `json:"type"`
	Subject string             `json:"subject"`
	SN      int64              `json:"sn"`
	Data    WSFutureTickerData `json:"data"`
}
