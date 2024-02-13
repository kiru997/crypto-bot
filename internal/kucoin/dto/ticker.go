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
