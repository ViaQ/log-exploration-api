package logscontroller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LogsController struct {
	logsProvider logs.LogsProvider
	log          *zap.Logger
}

func NewLogsController(log *zap.Logger, logsProvider logs.LogsProvider, router *gin.Engine) *LogsController {
	controller := &LogsController{
		log:          log,
		logsProvider: logsProvider,
	}
	r := router.Group("logs")
	r.GET("/", controller.GetAllLogs)
	//please enter time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE - +00:00]
	r.GET("timefilter/:startTime/:finishTime", controller.FilterLogsByTime)
	r.GET("indexfilter/:index", controller.FilterLogsByIndex)
	r.GET("podnamefilter/:podname", controller.FilterLogsByPodName)
	r.GET("multifilter/:podname/:namespace/:starttime/:finishtime", controller.FilterLogsMultipleParameters)
	return controller
}

func (controller *LogsController) FilterLogsByPodName(gctx *gin.Context) {

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

func (controller *LogsController) GetAllLogs(gctx *gin.Context) {

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

func (controller *LogsController) FilterLogsByIndex(gctx *gin.Context) {

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

func (controller *LogsController) FilterLogsByTime(gctx *gin.Context) {

	start := gctx.Params.ByName("startTime")
	finish := gctx.Params.ByName("finishTime")
	startTime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"Error": "Incorrect format: Please Enter Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]",
		})
		fmt.Println(err)
		return
	}
	finishTime, err := time.Parse(time.RFC3339Nano, finish)
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
func (controller *LogsController) FilterLogsMultipleParameters(gctx *gin.Context) {

	podName := gctx.Params.ByName("podname")
	namespace := gctx.Params.ByName("namespace")
	starttime := gctx.Params.ByName("starttime")
	finishtime := gctx.Params.ByName("finishtime")

	startTime, err := time.Parse(time.RFC3339Nano, starttime)
	if err != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"Error": "Incorrect format: Please Enter Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]",
		})
		fmt.Println(err)
		return
	}
	finishTime, err := time.Parse(time.RFC3339Nano, finishtime)
	if err != nil {
		gctx.JSON(http.StatusBadRequest,
			gin.H{"Error": "Incorrect format: Please Finish Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]"})
		fmt.Println(err)
		return

	}
	logsList, err := controller.logsProvider.FilterLogsMultipleParameters(podName, namespace, startTime, finishTime)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please Check Entered Parameters": err,
			})
			return
		}
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"An Error Occurred ": err,
		})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{
		"Logs": logsList, //return logs
	})
}
