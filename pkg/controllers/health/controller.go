package health

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewHealthController(router *gin.Engine) {
	router.GET("/health", HealthHandler)
}

func HealthHandler(gctx *gin.Context) {
	gctx.Header("Access-Control-Allow-Origin", "*")
	gctx.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers")

	gctx.JSON(http.StatusOK, gin.H{"Message": "Success"})
}
