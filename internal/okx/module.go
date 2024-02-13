package okx

import (
	"example.com/greetings/internal/okx/service"
	"go.uber.org/fx"
)

var Module = fx.Options(fx.Provide(service.NewFutureService))
