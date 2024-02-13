package kucoin

import (
	"example.com/greetings/internal/kucoin/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(service.NewKucoinSpotService),
	fx.Provide(service.NewKucoinFutureService),
)
