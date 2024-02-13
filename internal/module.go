package internal

import (
	"example.com/greetings/internal/binance"
	"example.com/greetings/internal/kucoin"
	"example.com/greetings/internal/mexc"
	"example.com/greetings/internal/okx"
	"example.com/greetings/internal/port/http"
	"example.com/greetings/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Options(
	kucoin.Module,
	mexc.Module,
	okx.Module,
	binance.Module,

	fx.Provide(service.NewCompareService),

	fx.Invoke(
		http.RegisterDiffController,
	),
)
