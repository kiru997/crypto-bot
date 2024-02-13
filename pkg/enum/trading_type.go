package enum

import (
	"bytes"
	"fmt"
)

type TradingType uint8

const (
	TradingTypeNone TradingType = iota
	TradingTypeSpot
	TradingTypeFuture
)

const (
	TradingTypeNameNone   = ""
	TradingTypeNameSpot   = "spot"
	TradingTypeNameFuture = "future"
)

var TradingTypeName = map[TradingType]string{
	TradingTypeNone:   "",
	TradingTypeSpot:   TradingTypeNameSpot,
	TradingTypeFuture: TradingTypeNameFuture,
}

var TradingTypeValue = func() map[string]TradingType {
	value := map[string]TradingType{}
	for k, v := range TradingTypeName {
		value[v] = k
		value[fmt.Sprintf("%v", k)] = k
	}

	return value
}()

func (e TradingType) MarshalJSON() ([]byte, error) {
	v, ok := TradingTypeName[e]
	if !ok {
		return []byte("\"\""), nil
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(v)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (e *TradingType) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	v, ok := TradingTypeValue[string(data)]
	if !ok {
		return fmt.Errorf("enum '%s' is not register, must be one of: %v", data, e.EnumDescriptions())
	}

	*e = v

	return nil
}

func (*TradingType) EnumDescriptions() []string {
	vals := []string{}

	for _, name := range TradingTypeName {
		vals = append(vals, name)
	}

	return vals
}
