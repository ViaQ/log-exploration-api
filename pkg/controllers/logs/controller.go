package logscontroller

import (
	"net/http"
	"strings"

	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/ViaQ/log-exploration-api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ResponseLogs struct {
	Error string   `json:"Error"`
	Logs  []string `json:"Logs"`
}

type LogsController struct {
	logsProvider logs.LogsProvider
	log          *zap.Logger
}

func NewLogsController(log *zap.Logger, logsProvider logs.LogsProvider, router *gin.Engine) *LogsController {
	controller := &LogsController{
		log:          log,
		logsProvider: logsProvider,
	}

	router.Use(middleware.AddHeader())
	r := router.Group("logs")
	r.Use(middleware.TokenHeader())
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

func initializeQueryParameters(gctx *gin.Context) logs.Parameters {
	var queryParams logs.Parameters
	err := gctx.Bind(&queryParams)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, &ResponseLogs{ //If error is not nil, an internal server error might have ocurred
			Error: err.Error() + ", an internal server error may have occurred",
			Logs:  nil,
		})
	}
	return queryParams
}

func emitFilteredLogs(gctx *gin.Context, logsList []string, err error) {
	if err != nil {
		if err.Error() == logs.NotFoundError().Error() { //If error is not nil, and logs are not nil, implies a user error has occurred
			gctx.JSON(http.StatusBadRequest, &ResponseLogs{
				Error: logs.NotFoundError().Error() + ", please check the input parameters",
				Logs:  logsList,
			})
			return
		} else if err.Error() == "context deadline exceeded" { // request timeout
			gctx.JSON(http.StatusRequestTimeout, &ResponseLogs{
				Error: err.Error() + ": request timeout",
				Logs:  logsList,
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, &ResponseLogs{ //If error is not nil and logs are not nil, implies an internal server error might have ocurred
				Error: err.Error() + ", an internal server error may have occurred",
				Logs:  logsList,
			})
			return
		}
	}
	gctx.JSON(http.StatusOK, gin.H{"Logs": logsList})
}

func (controller *LogsController) FilterPodLogs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	params.Namespace = gctx.Params.ByName("namespace")
	params.Podname = gctx.Params.ByName("podname")
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.FilterPodLogs(params)
	emitFilteredLogs(gctx, logsList, err)
}

func (controller *LogsController) Logs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.Logs(params)
	emitFilteredLogs(gctx, logsList, err)
}

func (controller *LogsController) FilterNamespaceLogs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	params.Namespace = gctx.Params.ByName("namespace")
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.FilterNamespaceLogs(params)
	emitFilteredLogs(gctx, logsList, err)
}

func (controller *LogsController) FilterContainerLogs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	params.Namespace = gctx.Params.ByName("namespace")
	params.ContainerName = gctx.Params.ByName("containername")
	params.Podname = gctx.Params.ByName("podname")
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.FilterContainerLogs(params)
	emitFilteredLogs(gctx, logsList, err)
}

func (controller *LogsController) FilterLabelLogs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	labels := gctx.Params.ByName("labels")
	labelsList := strings.Split(labels, ",") //split labels on "," to obtain a list of individual labels
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.FilterLabelLogs(params, labelsList)
	emitFilteredLogs(gctx, logsList, err)
}

func (controller *LogsController) FilterLogs(gctx *gin.Context) {
	params := initializeQueryParameters(gctx)
	params.Token = map[string]string{"Authorization": gctx.Request.Header["Authorization"][0]}
	logsList, err := controller.logsProvider.FilterLogs(params)
	emitFilteredLogs(gctx, logsList, err)
}
