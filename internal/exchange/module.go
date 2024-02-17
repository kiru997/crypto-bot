package exchange

import (
	"example.com/greetings/internal/exchange/binance"
	"example.com/greetings/internal/exchange/bitmart"
	"example.com/greetings/internal/exchange/bybit"
	"example.com/greetings/internal/exchange/gate"
	"example.com/greetings/internal/exchange/kucoin"
	"example.com/greetings/internal/exchange/mexc"
	"example.com/greetings/internal/exchange/okx"

	"go.uber.org/fx"
)

var Module = fx.Options(
	kucoin.Module,
	mexc.Module,
	binance.Module,
	bybit.Module,
	gate.Module,
	bitmart.Module,
	okx.Module,
)
