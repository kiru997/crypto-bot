package internal

import (
	"example.com/greetings/internal/exchange"
	"example.com/greetings/internal/port"
	"example.com/greetings/internal/service"

	"go.uber.org/fx"
)

var Module = fx.Options(
	exchange.Module,
	service.Module,
	port.Module,
)
