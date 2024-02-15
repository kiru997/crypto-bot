package kucoin

import (
	"example.com/greetings/internal/exchange/kucoin/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(service.NewSpotService),
	fx.Provide(service.NewFutureService),
)
