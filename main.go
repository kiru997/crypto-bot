package main

import (
	"context"
	"time"

	"example.com/greetings/internal"
	"example.com/greetings/pkg/cmdutil"
	"example.com/greetings/pkg/configs"
	"example.com/greetings/pkg/constants"
	"example.com/greetings/pkg/log"
	"example.com/greetings/pkg/routeutil"
	"example.com/greetings/pkg/transportutil"

	"example.com/greetings/internal/port/background"

	kfsdk "github.com/Kucoin/kucoin-futures-go-sdk"
	ksdk "github.com/Kucoin/kucoin-go-sdk"
	"github.com/gateio/gateapi-go/v6"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := configs.NewConfig("./config.yml")
	if err != nil {
		log.Panic("err load config:", log.Any("error", err))
	}

	compareConfig, err := configs.NewCompareConfig("./compare.config.json")
	if err != nil {
		log.Panic("err load config compare:", log.Any("error", err))
	}

	wrapLog := log.NewWrapLogger(log.GetLogLevel(cfg.LogLevel), cfg.Env == constants.EnvLocal)

	opts := []fx.Option{
		fx.Provide(func() context.Context {
			return ctx
		}),

		fx.Provide(func() *zap.Logger {
			return wrapLog.GetInstance()
		}),

		fx.Provide(func() *configs.AppConfig {
			return cfg
		}),

		fx.Provide(func() *configs.CompareConfig {
			return compareConfig
		}),

		fx.Provide(func(cgf *configs.AppConfig) *gateapi.APIClient {
			return gateapi.NewAPIClient(&gateapi.Configuration{
				BasePath: cgf.Gate.SpotAPIBaseURL,
			})
		}),

		fx.Provide(
			func() *ksdk.ApiService {
				return ksdk.NewApiService(
					ksdk.ApiKeyVersionOption(ksdk.ApiKeyVersionV2),
					ksdk.ApiBaseURIOption(cfg.Kucoin.SpotAPIBaseURL),
				)
			},
		),
		fx.Provide(
			func() *kfsdk.ApiService {
				return kfsdk.NewApiService(
					kfsdk.ApiKeyVersionOption(ksdk.ApiKeyVersionV2),
					kfsdk.ApiBaseURIOption(cfg.Kucoin.FutureAPIBaseURL),
				)
			},
		),

		internal.Module,

		fx.Provide(transportutil.InitHttpServer),
		fx.Provide(transportutil.InitGinEngine),

		fx.Provide(func(r *gin.Engine) *gin.RouterGroup {
			g := r.Group("")
			g.GET("api-docs", routeutil.ServingDocs)

			return g
		}),

		fx.Invoke(
			cmdutil.RunHTTPServer,
			background.RunWorker,
		),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: logger,
			}
		}),
	}

	err = fx.ValidateApp(opts...)
	if err != nil {
		log.Panic("err provide autowire", log.Any("error", err))
	}

	app := fx.New(opts...)

	startCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		log.Fatal("app.Start error", log.Any("error", err))
	}

	<-app.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		log.Fatal("app.Stop error", log.Any("error", err))
	}

}
