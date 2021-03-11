package logscontroller

import "github.com/gin-gonic/gin"

type LogsController interface {
	GetAllLogs(gctx *gin.Context)
	FilterLogsByTime(gctx *gin.Context)
	FilterLogsByIndex(gctx *gin.Context)
	FilterLogsByPodName(gctx *gin.Context)
}
