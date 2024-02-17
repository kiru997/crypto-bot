package dto

import "encoding/json"

type TickerMessageData struct {
	Symbol                  string    `json:"symbol"`
	LastPrice               float64   `json:"lastPrice"`
	RiseFallRate            float64   `json:"riseFallRate"`
	FairPrice               float64   `json:"fairPrice"`
	IndexPrice              float64   `json:"indexPrice"`
	Volume24                int       `json:"volume24"`
	Amount24                float64   `json:"amount24"`
	MaxBidPrice             float64   `json:"maxBidPrice"`
	MinAskPrice             float64   `json:"minAskPrice"`
	Lower24Price            float64   `json:"lower24Price"`
	High24Price             float64   `json:"high24Price"`
	Timestamp               int64     `json:"timestamp"`
	Bid1                    float64   `json:"bid1"`
	Ask1                    float64   `json:"ask1"`
	HoldVol                 int       `json:"holdVol"`
	RiseFallValue           float64   `json:"riseFallValue"`
	FundingRate             float64   `json:"fundingRate"`
	Zone                    string    `json:"zone"`
	RiseFallRates           []float64 `json:"riseFallRates"`
	RiseFallRatesOfTimezone []float64 `json:"riseFallRatesOfTimezone"`
}

type TickerMessage struct {
	Symbol  string            `json:"symbol"`
	Data    TickerMessageData `json:"data"`
	Channel string            `json:"channel"`
	Ts      int64             `json:"ts"`
}

type SpotTicker24hRes []*SpotTicker24hResItem

type SpotTicker24hResItem struct {
	Symbol             string      `json:"symbol"`
	PriceChange        string      `json:"priceChange"`
	PriceChangePercent json.Number `json:"priceChangePercent"`
	PrevClosePrice     string      `json:"prevClosePrice"`
	LastPrice          string      `json:"lastPrice"`
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
	Count              *int64      `json:"count"`
}

type WsSpotBookTickerMsg struct {
	C    string `json:"c"`
	Data struct {
		A         string      `json:"A"`
		B         string      `json:"B"`
		SellPrice string      `json:"a"`
		BuyPrice  json.Number `json:"b"`
	} `json:"d"`
	S string `json:"s"`
	T int64  `json:"t"`
}
