package dto

import "encoding/json"

type SpotTickerItem struct {
	Symbol         string      `json:"symbol"`
	LastPrice      json.Number `json:"last_price"`
	QuoteVolume24h string      `json:"quote_volume_24h"`
	BaseVolume24h  json.Number `json:"base_volume_24h"`
	High24h        string      `json:"high_24h"`
	Low24h         string      `json:"low_24h"`
	Open24h        string      `json:"open_24h"`
	Close24h       string      `json:"close_24h"`
	BestAsk        string      `json:"best_ask"`
	BestAskSize    string      `json:"best_ask_size"`
	BestBid        string      `json:"best_bid"`
	BestBidSize    string      `json:"best_bid_size"`
	Fluctuation    string      `json:"fluctuation"`
	URL            string      `json:"url"`
	Timestamp      int64       `json:"timestamp"`
}

type SpotTickerResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Trace   string `json:"trace"`
	Data    struct {
		Tickers []*SpotTickerItem `json:"tickers"`
	} `json:"data"`
}

type WSSpotTickerResponseItem struct {
	BaseVolume24h json.Number `json:"base_volume_24h"`
	High24h       string      `json:"high_24h"`
	LastPrice     json.Number `json:"last_price"`
	Low24h        string      `json:"low_24h"`
	Open24h       string      `json:"open_24h"`
	Timestamp     int64       `json:"s_t"`
	Symbol        string      `json:"symbol"`
}

type WSSpotTickerResponse struct {
	Data  []WSSpotTickerResponseItem `json:"data"`
	Table string                     `json:"table"`
}

type WSFuturesTickerResponseData struct {
	Symbol    string      `json:"symbol"`
	LastPrice json.Number `json:"last_price"`
	Volume24  string      `json:"volume_24"`
	Range     string      `json:"range"`
	FairPrice string      `json:"fair_price"`
	AskPrice  string      `json:"ask_price"`
	AskVol    string      `json:"ask_vol"`
	BidPrice  string      `json:"bid_price"`
	BidVol    string      `json:"bid_vol"`
}

type WSFuturesTickerResponse struct {
	Group string                      `json:"group"`
	Data  WSFuturesTickerResponseData `json:"data"`
}

type WSSpotBookTickerMsg struct {
	OP   string   `json:"op"`
	Args []string `json:"args"`
}

type WSFutureTickerMsg struct {
	Action string   `json:"action"`
	Args   []string `json:"args"`
}

type WSFutureKlineResponseDataItem struct {
	Open   string      `json:"o"`
	High   string      `json:"h"`
	Low    string      `json:"l"`
	Close  json.Number `json:"c"`
	Volume string      `json:"v"`
	TS     int64       `json:"ts"`
}

type WSFutureKlineResponseData struct {
	Symbol string                          `json:"symbol"`
	Items  []WSFutureKlineResponseDataItem `json:"items"`
}

type WSFutureKlineResponse struct {
	Group string                    `json:"group"`
	Data  WSFutureKlineResponseData `json:"data"`
}
