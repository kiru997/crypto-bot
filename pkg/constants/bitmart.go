package constants

import "time"

const (
	BitmartWSOpSubscribe            = "subscribe"
	BitmartWSOpUnSubscribe          = "unsubscribe"
	BitmartWSActionSubscribe        = "subscribe"
	BitmartWSActionUnSubscribe      = "unsubscribe"
	BitmartWSActionPing             = "ping"
	BitmartWSActionPong             = "pong"
	BitmartWSSpotTickerPrefix       = "spot/ticker:"
	BitmartWSFutureTicker           = "futures/ticker"
	BitmartWSFutureKlineBin1mPrefix = "futures/klineBin1m:"

	// 10 req/second
	BitmartWSRequestSleep = time.Second / 10
)
