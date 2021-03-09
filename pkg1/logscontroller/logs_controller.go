package logscontroller

import "github.com/gin-gonic/gin"

type LogsController interface {
	GetAllLogs() gin.HandlerFunc
	FilterLogsByTime() gin.HandlerFunc
	FilterLogsByIndex() gin.HandlerFunc
	FilterLogsByPodName() gin.HandlerFunc
}
