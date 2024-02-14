package enum

import (
	"bytes"
	"fmt"
)

type ExchangeType uint8

const (
	ExchangeTypeNone ExchangeType = iota
	ExchangeTypeKucoin
	ExchangeTypeKucoinFuture
	ExchangeTypeMexc
	ExchangeTypeMexcFuture
	ExchangeTypeOkx
	ExchangeTypeBinance
	ExchangeTypeBinanceFuture
	ExchangeTypeBybit
	ExchangeTypeBybitFuture
	ExchangeTypeGate
	ExchangeTypeGateFuture
)

const (
	ExchangeTypeNameNone          = ""
	ExchangeTypeNameKucoin        = "kucoin"
	ExchangeTypeNameKucoinFuture  = "kucoin_future"
	ExchangeTypeNameMexc          = "mexc"
	ExchangeTypeNameMexcFuture    = "mexc_future"
	ExchangeTypeNameOkx           = "okx"
	ExchangeTypeNameBinance       = "binance"
	ExchangeTypeNameBinanceFuture = "binance_future"
	ExchangeTypeNameBybit         = "bybit"
	ExchangeTypeNameBybitFuture   = "bybit_future"
	ExchangeTypeNameGate          = "gate"
	ExchangeTypeNameGateFuture    = "gate_future"
)

var ExchangeTypeName = map[ExchangeType]string{
	ExchangeTypeNone:          "",
	ExchangeTypeKucoin:        ExchangeTypeNameKucoin,
	ExchangeTypeKucoinFuture:  ExchangeTypeNameKucoinFuture,
	ExchangeTypeMexc:          ExchangeTypeNameMexc,
	ExchangeTypeMexcFuture:    ExchangeTypeNameMexcFuture,
	ExchangeTypeOkx:           ExchangeTypeNameOkx,
	ExchangeTypeBinance:       ExchangeTypeNameBinance,
	ExchangeTypeBinanceFuture: ExchangeTypeNameBinanceFuture,
	ExchangeTypeBybit:         ExchangeTypeNameBybit,
	ExchangeTypeBybitFuture:   ExchangeTypeNameBybitFuture,
	ExchangeTypeGate:          ExchangeTypeNameGate,
	ExchangeTypeGateFuture:    ExchangeTypeNameGateFuture,
}

var ExchangeTypeValue = func() map[string]ExchangeType {
	value := map[string]ExchangeType{}
	for k, v := range ExchangeTypeName {
		value[v] = k
		value[fmt.Sprintf("%v", k)] = k
	}

	return value
}()

func (e ExchangeType) MarshalJSON() ([]byte, error) {
	v, ok := ExchangeTypeName[e]
	if !ok {
		return []byte("\"\""), nil
	}

	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(v)
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (e *ExchangeType) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	v, ok := ExchangeTypeValue[string(data)]
	if !ok {
		return fmt.Errorf("enum '%s' is not register, must be one of: %v", data, e.EnumDescriptions())
	}

	*e = v

	return nil
}

func (*ExchangeType) EnumDescriptions() []string {
	vals := []string{}

	for _, name := range ExchangeTypeName {
		vals = append(vals, name)
	}

	return vals
}
