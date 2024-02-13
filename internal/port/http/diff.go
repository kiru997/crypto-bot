package http

import (
	"example.com/greetings/internal/service"
	"example.com/greetings/pkg/configs"
	"github.com/gin-gonic/gin"
)

type DiffController interface {
	SubcribeSymbols(c *gin.Context)
}

type diffController struct {
	cfg *configs.AppConfig
	sv  service.CompareService
}

func RegisterDiffController(
	r *gin.RouterGroup,
	cfg *configs.AppConfig,
	sv service.CompareService,
) {
	g := r.Group("diff")

	var c DiffController = &diffController{
		cfg: cfg,
		sv:  sv,
	}

	g.GET("/subcribe-symbols", c.SubcribeSymbols)
}
