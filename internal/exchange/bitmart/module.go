package bitmart

import (
	"example.com/greetings/internal/exchange/bitmart/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(service.NewSpotService),
	fx.Provide(service.NewFutureService),
)
