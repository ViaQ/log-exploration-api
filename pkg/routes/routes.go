package routes
import(
	"github.com/gin-gonic/gin"
	 "logexplorationapi/pkg/controllers"
	"github.com/elastic/go-elasticsearch/v7"
	"logexplorationapi/pkg/elastic"
)

func SetUpRouter() *gin.Engine{
	r := gin.Default() //initialise

	r.Use(AddESClientToContext())
	//	endpoint to get all logs
	r.GET("/", controllers.GetAllLogs)

	//endpoint to get infrastructure logs
	r.GET("/infra", controllers.GetInfrastructureLogs)

	//endpoint to get application logs
	r.GET("/app", controllers.GetApplicationLogs)

	//endpoint to get audit logs
	r.GET("/audit", controllers.GetAuditLogs)

	//endpoint to filter logs by start and finish time - please enter time in the following format- HH:MM:SS
	r.GET("/filter/:startTime/:finishTime", controllers.FilterByTime)

	return r

}
func AddESClientToContext() gin.HandlerFunc {
	//add ESClient to context
	esClient:= elastic.InitializeElasticSearchClient()
	return func(c *gin.Context) {
		c.Set("esClient", esClient)
		c.Next()
	}
}
