package logscontroller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"logexplorationapi/pkg/logs"
	"net/http"
	"time"
)

type logsController struct {
	logsProvider logs.LogsProvider
}

func NewLogsController(logsProvider logs.LogsProvider, router *gin.Engine) {
	controller := &logsController{logsProvider: logsProvider}
	router.GET("/", controller.GetAllLogs)
	//please enter time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE - +00:00]
	router.GET("timefilter/:startTime/:finishTime", controller.FilterLogsByTime)
	router.GET("indexfilter/:index", controller.FilterLogsByIndex)
	router.GET("podnamefilter/:podname", controller.FilterLogsByPodName)

}

func (controller *logsController) FilterLogsByPodName(gctx *gin.Context) {

	podName := gctx.Params.ByName("podname")
	logsList, err := controller.logsProvider.FilterByPodName(podName)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			//If Error is not nil, and logs is nil, A user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Invalid Podname Entered ": err,
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If Error is not nil and logs is not nil, an internal error might have ocurred
				"An Error Occurred": err,
			})
			return
		}
	}
	gctx.JSON(http.StatusOK, gin.H{
		"Logs": logsList, //return logs
	})
	return

}

func (controller *logsController) GetAllLogs(gctx *gin.Context) {

	logsList, err := controller.logsProvider.GetAllLogs()

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			gctx.JSON(http.StatusBadRequest, gin.H{
				"An Error Occurred ": err,
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"An Error Occurred ": err,
			})
			return
		}
	}
	gctx.JSON(http.StatusOK, gin.H{
		"Logs": logsList, //return logs
	})
}

func (controller *logsController) FilterLogsByIndex(gctx *gin.Context) {

	index := gctx.Params.ByName("index")
	logsList, err := controller.logsProvider.FilterByIndex(index)
	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Invalid Index Entered ": err,
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"An Error Occurred ": err,
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"Logs": logsList, //return logs
	})
}

func (controller *logsController) FilterLogsByTime(gctx *gin.Context) {

	start := gctx.Params.ByName("startTime")
	finish := gctx.Params.ByName("finishTime")

	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"Error": "Incorrect format: Please Enter Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]",
		})
		fmt.Println(err)
		return
	}
	finishTime, err := time.Parse(time.RFC3339, finish)
	if err != nil {
		gctx.JSON(http.StatusBadRequest,
			gin.H{"Error": "Incorrect format: Please Finish Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]"})
		fmt.Println(err)
		return

	}

	logsList, err := controller.logsProvider.FilterByTime(startTime, finishTime)
	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please Enter a Valid timeStamp ": err,
			})
			return
		}
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"An Error Occurred": err,
		})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{
		"Logs": logsList, //return logs
	})
}
