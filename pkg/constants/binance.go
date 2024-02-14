package constants

import "time"

const (
	BinanceWSMethodSubscribe   = "SUBSCRIBE"
	BinanceWSMethodUnSubscribe = "UNSUBSCRIBE"

	// 10 req/second
	BinanceWSRequestSleep = time.Second / 5
)
