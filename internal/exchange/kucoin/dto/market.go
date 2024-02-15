package dto

type FutureMarketLv2Message struct {
	Subject string `json:"subject"`
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Data    struct {
		Sequence  int    `json:"sequence"`
		Change    string `json:"change"`
		Timestamp int64  `json:"timestamp"`
	} `json:"data"`
}
