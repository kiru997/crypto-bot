package dto

import (
	"time"

	"example.com/greetings/pkg/enum"
)

type CompareSymbolNotiExchangeItem struct {
	ExchangeType enum.ExchangeType
	Price        float64
	LastPriceAt  time.Time
	LastNotiAt   time.Time
	Percent      float64
}

type CompareSymbolNotiItem struct {
	Symbol       string
	SpotPrices   []*CompareSymbolNotiExchangeItem
	FuturePrices []*CompareSymbolNotiExchangeItem
}

type ComparePriceChanMsg struct {
	ExchangeType enum.ExchangeType
	TradingType  enum.TradingType
	Symbol       string
	Price        float64
	At           time.Time
	ConID        string
}
