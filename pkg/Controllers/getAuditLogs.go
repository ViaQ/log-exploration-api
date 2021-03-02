package Controllers

import (
	"context"
	"fmt"

	"hello/pkg/Utils"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func GetAuditLogs(c *gin.Context) {

	esClient, ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	//fetch logs from index audit-000001
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("audit-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}
	var logs []string
	logs = Utils.GetRelevantLogs(searchResult)

	c.JSON(200, gin.H{
		"Audit Logs": logs, //return logs
	})

}
