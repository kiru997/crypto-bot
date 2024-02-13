package dto

import "time"

type TickerPrice struct {
	Symbol string
	Time   time.Time
	Price  float64
}
