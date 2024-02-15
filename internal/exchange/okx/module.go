package okx

import (
	"example.com/greetings/internal/exchange/okx/service"
	"go.uber.org/fx"
)

var Module = fx.Options(fx.Provide(service.NewSpotService))
