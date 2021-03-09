package logscontroller

import (
	"github.com/gin-gonic/gin"
	"logexplorationapi/pkg1/logs"
	"regexp"
)

type logsController struct {
	logsProvider logs.LogsProvider
}

func NewLogsController(logsProvider logs.LogsProvider) *gin.Engine {
	controller := &logsController{logsProvider: logsProvider}
	router := gin.Default()
	router.GET("/", controller.GetAllLogs())
	//please enter time in the following format YYYY-MM-DDTHH:MM::SS
	router.GET("timefilter/:startTime/:finishTime", controller.FilterLogsByTime())
	router.GET("indexfilter/:index", controller.FilterLogsByIndex())
	router.GET("podnamefilter/:podname", controller.FilterLogsByPodName())
	return router
}

func (controller *logsController) FilterLogsByPodName() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		var logs []string
		podName := gctx.Params.ByName("podname")
		logs = controller.logsProvider.FilterByPodName(podName)
		gctx.JSON(200, gin.H{
			"Logs": logs, //return logs
		})
	}
}

func (controller *logsController) GetAllLogs() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		var logs []string
		logs = controller.logsProvider.GetAllLogs()
		gctx.JSON(200, gin.H{
			"Logs": logs, //return logs
		})
	}

}
func (controller *logsController) FilterLogsByIndex() gin.HandlerFunc {

	return func(gctx *gin.Context) {
		index := gctx.Params.ByName("index")
		var logs []string
		logs = controller.logsProvider.FilterByIndex(index)
		gctx.JSON(200, gin.H{
			"Logs": logs, //return logs
		})
	}
}

func (controller *logsController) FilterLogsByTime() gin.HandlerFunc {

	return func(gctx *gin.Context) {
		startTime := gctx.Params.ByName("startTime")
		finishTime := gctx.Params.ByName("finishTime")

		pattern := "[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]"
		checkStartTime, _ := regexp.MatchString(pattern, startTime)
		checkFinishTime, _ := regexp.MatchString(pattern, finishTime)
		var logs []string
		if checkStartTime && checkFinishTime {
			logs = controller.logsProvider.FilterByTime(startTime, finishTime)
		} else {
			logs = append(logs, "Incorrect format: Please Enter Time in the following format YYYY-MM-DDTHH:MM:SS")
		}
		gctx.JSON(200, gin.H{
			"Logs": logs, //return logs
		})
	}
}
