package Routes

import (
	"hello/pkg/Controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	//r := gin.Default()
	baseRoute := r.Group("/")
	{
		//	endpoint to get all logs
		baseRoute.GET("/", Controllers.GetAllLogs)

		//endpoint to get infrastructure logs
		baseRoute.GET("/infra", Controllers.GetInfrastructureLogs)

		//endpoint to get application logs
		baseRoute.GET("/app", Controllers.GetApplicationLogs)

		//endpoint to get audit logs
		baseRoute.GET("/audit", Controllers.GetAuditLogs)

		//endpoint to filter logs by start and finish time - please enter time in the following format- HH:MM:SS
		baseRoute.GET("/filter/:startTime/:finishTime", Controllers.FilterByTime)
	}

	return r
}
