package dto

import "example.com/greetings/internal/kucoin/enum"

type WSWriteMessage struct {
	ID             int64               `json:"id"`
	Type           enum.WSWriteMsgType `json:"type"`
	Topic          string              `json:"topic,omitempty"`
	PrivateChannel bool                `json:"privateChannel,omitempty"`
	Response       bool                `json:"response,omitempty"`
}
