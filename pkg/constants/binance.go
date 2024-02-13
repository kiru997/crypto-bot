package constants

import "time"

const (
	BinanceWSMethodSubcribe   = "SUBSCRIBE"
	BinanceWSMethodUnSubcribe = "UNSUBSCRIBE"

	// 10 req/second
	BinanceWSRequestSleep = time.Second / 5
)
