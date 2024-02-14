package gate

import (
	"example.com/greetings/internal/gate/service"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(service.NewSpotService),
	fx.Provide(service.NewFutureService),
)
