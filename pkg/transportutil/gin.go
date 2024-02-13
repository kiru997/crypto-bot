package transportutil

import (
	"net/http"
	"sync"
	"time"

	"example.com/greetings/pkg/configs"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var publicEndpoint = map[string]struct{}{
	"/api-docs": {},
	"/metrics":  {},
	"/healthz":  {},
}

var l sync.RWMutex

func RegisterPublicEndpoint(path string) {
	l.Lock()
	defer l.Unlock()

	publicEndpoint[path] = struct{}{}
}

func InitHttpServer(cfg *configs.AppConfig, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    cfg.Port,
		Handler: router,
	}
}

func InitGinEngine(cfg *configs.AppConfig, logger *zap.Logger) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	pprof.Register(r)

	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, "ok")
	})

	r.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "DeviceID", "Accept-Language"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		AllowAllOrigins:  true,
	}))
	r.Use(requestid.New())

	return r
}
