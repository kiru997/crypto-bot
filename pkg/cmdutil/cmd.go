package cmdutil

import (
	"context"
	"net/http"
	"time"

	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/log"
	"go.uber.org/fx"
)

func RunHTTPServer(
	lifecycle fx.Lifecycle,
	cfg *configs.AppConfig,
	srv *http.Server,
) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						log.Fatal("r.Run", log.String("port", cfg.Port), log.Any("error", err))
					}
				}()

				log.Info("HTTP server is running", log.String("port", cfg.Port), log.String("env", cfg.Env), log.Bool("debug", cfg.Debug))
				return nil
			},
			OnStop: func(ctx context.Context) error {
				log.Info("shutting down server...")

				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				return srv.Shutdown(ctx)
			},
		},
	)
}
