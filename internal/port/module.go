package port

import (
	"example.com/greetings/internal/port/http"

	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Invoke(
		http.RegisterDiffController,
	),
)
