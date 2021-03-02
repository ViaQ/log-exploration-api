package main

import (
	"hello/pkg/Config"
	"hello/pkg/Routes"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func main() {

	var esClient *elasticsearch.Client

	//esClient = InitializeElasticSearchClient(esAddress, esCert, esKey, esClient)
	esClient = Config.InitializeElasticSearchClient()

	r := gin.Default() //initialise

	r.Use(Config.AddESClientToContext(esClient))

	//fmt.Println(reflect.TypeOf(r))

	router := Routes.SetupRouter(r)

	router.Run(":8080") //run server on port 8080
}
