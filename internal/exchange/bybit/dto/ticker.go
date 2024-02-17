package dto

import "encoding/json"

type WSSpotTickerMessageData struct {
	Symbol        string      `json:"symbol"`
	LastPrice     json.Number `json:"lastPrice"`
	HighPrice24h  string      `json:"highPrice24h"`
	LowPrice24h   string      `json:"lowPrice24h"`
	PrevPrice24h  string      `json:"prevPrice24h"`
	Volume24h     json.Number `json:"volume24h"`
	Turnover24h   string      `json:"turnover24h"`
	Price24hPcnt  string      `json:"price24hPcnt"`
	USDIndexPrice string      `json:"usdIndexPrice"`
}

type WSSpotTickerMessage struct {
	Topic string                  `json:"topic"`
	Ts    int64                   `json:"ts"`
	Type  string                  `json:"type"`
	Cs    int64                   `json:"cs"`
	Data  WSSpotTickerMessageData `json:"data"`
}

type WSFutureTickerMessageData struct {
	Symbol            string      `json:"symbol"`
	Price24hPcnt      string      `json:"price24hPcnt"`
	MarkPrice         string      `json:"markPrice"`
	IndexPrice        string      `json:"indexPrice"`
	OpenInterest      string      `json:"openInterest"`
	OpenInterestValue string      `json:"openInterestValue"`
	FundingRate       string      `json:"fundingRate"`
	Bid1Price         json.Number `json:"bid1Price"`
	Bid1Size          string      `json:"bid1Size"`
	Ask1Price         string      `json:"ask1Price"`
	Ask1Size          string      `json:"ask1Size"`
}

type WSFutureTickerMessage struct {
	Topic string                    `json:"topic"`
	Type  string                    `json:"type"`
	Data  WSFutureTickerMessageData `json:"data"`
	Cs    int64                     `json:"cs"`
	Ts    int64                     `json:"ts"`
}

type MarketTickerItem struct {
	Symbol                 string      `json:"symbol"`
	LastPrice              json.Number `json:"lastPrice"`
	IndexPrice             string      `json:"indexPrice"`
	MarkPrice              string      `json:"markPrice"`
	PrevPrice24h           string      `json:"prevPrice24h"`
	Price24hPcnt           json.Number `json:"price24hPcnt"`
	HighPrice24h           string      `json:"highPrice24h"`
	LowPrice24h            string      `json:"lowPrice24h"`
	PrevPrice1h            string      `json:"prevPrice1h"`
	OpenInterest           string      `json:"openInterest"`
	OpenInterestValue      string      `json:"openInterestValue"`
	Turnover24h            string      `json:"turnover24h"`
	Volume24h              json.Number `json:"volume24h"`
	FundingRate            string      `json:"fundingRate"`
	NextFundingTime        string      `json:"nextFundingTime"`
	PredictedDeliveryPrice string      `json:"predictedDeliveryPrice"`
	BasisRate              string      `json:"basisRate"`
	DeliveryFeeRate        string      `json:"deliveryFeeRate"`
	DeliveryTime           string      `json:"deliveryTime"`
	Ask1Size               string      `json:"ask1Size"`
	Bid1Price              string      `json:"bid1Price"`
	Ask1Price              string      `json:"ask1Price"`
	Bid1Size               string      `json:"bid1Size"`
	Basis                  string      `json:"basis"`
}

type MarketTickers struct {
	Category string              `json:"category"`
	List     []*MarketTickerItem `json:"list"`
}
