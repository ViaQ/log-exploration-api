package logscontroller

import (
	"net/http"
	"strings"

	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type LogsController struct {
	logsProvider logs.LogsProvider
	log          *zap.Logger
}

func addHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Header("Access-Control-Allow-Origin", "*")
		gctx.Next()
	}
}

func NewLogsController(log *zap.Logger, logsProvider logs.LogsProvider, router *gin.Engine) *LogsController {
	controller := &LogsController{
		log:          log,
		logsProvider: logsProvider,
	}

	router.Use(addHeader())
	r := router.Group("logs")
	r.GET("/filter", controller.FilterLogs)
	r.GET("/namespace/:namespace", controller.FilterNamespaceLogs)
	r.GET("/namespace/:namespace/pod/:podname", controller.FilterPodLogs)
	r.GET("/namespace/:namespace/pod/:podname/container/:containername", controller.FilterContainerLogs)
	r.GET("", controller.Logs)
	r.GET("/logs/namespace/:namespace/:entity/:entity_name", controller.FilterEntityLogs)
	r.GET("/logs_by_labels/:labels", controller.FilterLabelLogs)
	return controller
}

func (controller *LogsController) FilterEntityLogs(gctx *gin.Context) {

	gctx.JSON(http.StatusOK, gin.H{"Logs": "To Be Implemented"})

}

func (controller *LogsController) FilterPodLogs(gctx *gin.Context) {

	var params logs.Parameters

	params.Namespace = gctx.Params.ByName("namespace")
	params.Podname = gctx.Params.ByName("podname")

	err := gctx.Bind(&params)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have occurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.FilterPodLogs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidLimit().Error() || err.Error() == logs.InvalidTimeStamp().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}
func (controller *LogsController) Logs(gctx *gin.Context) {

	var params logs.Parameters

	err := gctx.Bind(&params)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have ocurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.Logs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidTimeStamp().Error() || err.Error() == logs.InvalidLimit().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}

func (controller *LogsController) FilterNamespaceLogs(gctx *gin.Context) {

	var params logs.Parameters

	params.Namespace = gctx.Params.ByName("namespace")

	err := gctx.Bind(&params)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have ocurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.FilterNamespaceLogs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidTimeStamp().Error() || err.Error() == logs.InvalidLimit().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}

func (controller *LogsController) FilterContainerLogs(gctx *gin.Context) {

	var params logs.Parameters

	params.Namespace = gctx.Params.ByName("namespace")
	params.ContainerName = gctx.Params.ByName("containername")
	params.Podname = gctx.Params.ByName("podname")
	err := gctx.Bind(&params)

	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have ocurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.FilterContainerLogs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidTimeStamp().Error() || err.Error() == logs.InvalidLimit().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}

func (controller *LogsController) FilterLabelLogs(gctx *gin.Context) {

	var params logs.Parameters

	labels := gctx.Params.ByName("labels")
	labelsList := strings.Split(labels, ",") //split labels on "," to obtain a list of individual labels
	err := gctx.Bind(&params)

	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have ocurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.FilterLabelLogs(params, labelsList)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidTimeStamp().Error() || err.Error() == logs.InvalidLimit().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}

func (controller *LogsController) FilterLogs(gctx *gin.Context) {

	var params logs.Parameters
	err := gctx.Bind(&params)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil, an internal server error might have ocurred
			"An error occurred": []string{err.Error()},
		})
	}

	logsList, err := controller.logsProvider.FilterLogs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() || err.Error() == logs.InvalidTimeStamp().Error() || err.Error() == logs.InvalidLimit().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please check the input parameters": []string{err.Error()},
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				"An error occurred": []string{err.Error()},
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})

}
