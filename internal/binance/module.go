package binance

import (
	"example.com/greetings/internal/binance/service"
	"go.uber.org/fx"
)

var Module = fx.Options(fx.Provide(service.NewFutureService))
