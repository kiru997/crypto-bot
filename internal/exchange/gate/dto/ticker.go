package dto

import "encoding/json"

type WSSpotTickerMessageResult struct {
	CurrencyPair     string      `json:"currency_pair"`
	Last             json.Number `json:"last"`
	LowestAsk        string      `json:"lowest_ask"`
	HighestBid       string      `json:"highest_bid"`
	ChangePercentage string      `json:"change_percentage"`
	BaseVolume       string      `json:"base_volume"`
	QuoteVolume      string      `json:"quote_volume"`
	High24h          string      `json:"high_24h"`
	Low24h           string      `json:"low_24h"`
}

type WSSpotTickerMessage struct {
	Time    int64                     `json:"time"`
	TimeMs  int64                     `json:"time_ms"`
	Channel string                    `json:"channel"`
	Event   string                    `json:"event"`
	Result  WSSpotTickerMessageResult `json:"result"`
}

type WSFutureTickerMessageResult struct {
	Contract              string      `json:"contract"`
	Last                  json.Number `json:"last"`
	ChangePercentage      string      `json:"change_percentage"`
	TotalSize             string      `json:"total_size"`
	Volume24h             string      `json:"volume_24h"`
	Volume24hBase         string      `json:"volume_24h_base"`
	Volume24hQuote        string      `json:"volume_24h_quote"`
	Volume24hSettle       string      `json:"volume_24h_settle"`
	MarkPrice             string      `json:"mark_price"`
	FundingRate           string      `json:"funding_rate"`
	FundingRateIndicative string      `json:"funding_rate_indicative"`
	IndexPrice            string      `json:"index_price"`
	QuantoBaseRate        string      `json:"quanto_base_rate"`
	Low24h                string      `json:"low_24h"`
	High24h               string      `json:"high_24h"`
}

type WSFutureTickerMessage struct {
	Time    int64                         `json:"time"`
	TimeMs  int64                         `json:"time_ms"`
	Channel string                        `json:"channel"`
	Event   string                        `json:"event"`
	Result  []WSFutureTickerMessageResult `json:"result"`
}
