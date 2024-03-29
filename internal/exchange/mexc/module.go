package mexc

import (
	"example.com/greetings/internal/exchange/mexc/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(service.NewSpotService),
	fx.Provide(service.NewFutureService),
)
