package dto

import "encoding/json"

type TickerData struct {
	InstId  string      `json:"instId"`
	IdxPx   json.Number `json:"idxPx"`
	Open24h string      `json:"open24h"`
	High24h string      `json:"high24h"`
	Low24h  string      `json:"low24h"`
	SodUtc0 string      `json:"sodUtc0"`
	SodUtc8 string      `json:"sodUtc8"`
	Ts      json.Number `json:"ts"`
}

type TickerArgs struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type TickerMessage struct {
	Arg  TickerArgs   `json:"arg"`
	Data []TickerData `json:"data"`
}
