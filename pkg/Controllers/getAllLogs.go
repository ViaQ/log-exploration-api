package Controllers

import (
	"context"
	"fmt"

	//"reflect"

	"hello/pkg/Utils"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func GetAllLogs(c *gin.Context) {
	esClient, ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("app-000001", "infra-000001", "audit-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err == nil {
		fmt.Println("Error getting response: %s", err)
	}

	var logs []string // create a slice of type string to append logs to

	logs = Utils.GetRelevantLogs(searchResult)
	fmt.Println(searchResult)

	c.JSON(200, gin.H{
		"All Logs": logs, //return logs
	})

}
