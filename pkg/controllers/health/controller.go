package health

import (
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/ViaQ/log-exploration-api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthController struct {
	healthProvider logs.LogsProvider
}

func NewHealthController(router *gin.Engine, logsProvider logs.LogsProvider) {
	healthController := &HealthController{
		healthProvider: logsProvider,
	}
	router.Use(middleware.AddHeader())
	r := router.Group("")
	r.GET("/health", healthController.HealthHandler)
	r.GET("/ready", healthController.ReadinessHandler)
}

func (healthController *HealthController) HealthHandler(gctx *gin.Context) {
	gctx.JSON(http.StatusOK, gin.H{"Message": "Success"})
}

func (healthController *HealthController) ReadinessHandler(gctx *gin.Context) {
	checkReadiness := healthController.healthProvider.CheckReadiness()
	if checkReadiness == false {
		gctx.JSON(http.StatusBadRequest, gin.H{"Message": "failed to connect to esClient"})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"Message": "Success"})
}
