package logscontroller

import (
	"net/http"

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
	//please enter time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE - +00:00]
	r.GET("/filter", controller.FilterLogs)
	return controller
}

func (controller *LogsController) FilterLogs(gctx *gin.Context) {

	var params logs.Parameters
	err := gctx.BindJSON(&params)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{ //If Error is not nil and logs is not nil, an internal error might have ocurred
			"An Error Occurred": err,
		})
	}
	logsList, err := controller.logsProvider.FilterLogs(params)

	if err != nil {
		if err.Error() == logs.NotFoundError().Error() {
			//If Error is not nil, and logs is nil, A user error has occurred
			gctx.JSON(http.StatusBadRequest, gin.H{
				"Please Check input parameters": err,
			})
			return
		} else {
			gctx.JSON(http.StatusInternalServerError, gin.H{ //If Error is not nil and logs is not nil, an internal error might have ocurred
				"An Error Occurred": err,
			})
			return
		}
	}

	gctx.JSON(http.StatusOK, gin.H{"logs": logsList})

}
