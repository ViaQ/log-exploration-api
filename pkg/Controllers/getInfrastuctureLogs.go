package Controllers

import (
	"context"
	"fmt"

	"hello/pkg/Utils"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func GetInfrastructureLogs(c *gin.Context) {

	esClient, ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	//fetch logs from index infra-000001
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("infra-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	if err != nil {
		fmt.Println(err)
	}

	var logs []string

	logs = Utils.GetRelevantLogs(searchResult) // create a slice of type string to append logs to

	c.JSON(200, gin.H{
		"Infra Logs": logs, //return logs
	})

}
