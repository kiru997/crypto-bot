package enum

import (
	"bytes"
	"fmt"
)

type WSWriteMsgType uint8

const (
	WSWriteMsgTypeNone WSWriteMsgType = iota
	WSWriteMsgTypePing
	WSWriteMsgTypeSubscribe
	WSWriteMsgTypeUnSubscribe
	WSWriteMsgTypeMessage
)

var WSWriteMsgTypeName = map[WSWriteMsgType]string{
	WSWriteMsgTypeNone:        "",
	WSWriteMsgTypePing:        "ping",
	WSWriteMsgTypeSubscribe:   "subscribe",
	WSWriteMsgTypeUnSubscribe: "unsubscribe",
	WSWriteMsgTypeMessage:     "message",
}

var WSWriteMsgTypeValue = func() map[string]WSWriteMsgType {
	value := map[string]WSWriteMsgType{}
	for k, v := range WSWriteMsgTypeName {
		value[v] = k
		value[fmt.Sprintf("%v", k)] = k
	}

	return value
}()

func (e WSWriteMsgType) MarshalJSON() ([]byte, error) {
	v, ok := WSWriteMsgTypeName[e]
	if !ok {
		return []byte("\"\""), nil
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(v)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (e *WSWriteMsgType) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	v, ok := WSWriteMsgTypeValue[string(data)]
	if !ok {
		return fmt.Errorf("enum '%s' is not register, must be one of: %v", data, e.EnumDescriptions())
	}

	*e = v

	return nil
}

func (*WSWriteMsgType) EnumDescriptions() []string {
	vals := []string{}

	for _, name := range WSWriteMsgTypeName {
		vals = append(vals, name)
	}

	return vals
}
