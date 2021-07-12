package health

import (
	"github.com/ViaQ/log-exploration-api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewHealthController(router *gin.Engine) {
	router.GET("/health", HealthHandler)
}

func HealthHandler(gctx *gin.Context) {
	middleware.AddHeader()
	gctx.JSON(http.StatusOK, gin.H{"Message": "Success"})
}
