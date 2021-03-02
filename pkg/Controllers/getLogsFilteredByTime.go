package Controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func FilterByTime(c *gin.Context) {

	startTime := c.Params.ByName("startTime")
	finishTime := c.Params.ByName("finishTime")

	fmt.Println(startTime)
	fmt.Println(finishTime)

	esClient, ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	var result map[string]interface{}
	var logs []string // create a slice of type string to append logs to
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("infra-000001", "app-000001", "audit-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}
	json.NewDecoder(searchResult.Body).Decode(&result)
	fmt.Println(result)
	if err != nil {
		fmt.Println(err)
	}

	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) { //iterate through logs and check if timestamp lies between start and end times
		log := fmt.Sprintf("%v", hit)
		index := strings.Index(log, "@timestamp")
		fmt.Println(index)
		time := log[index+22 : index+30]
		if time >= startTime && time <= finishTime {
			logs = append(logs, log)
			fmt.Println(hit, "\n")
			fmt.Println()
		}
	}

	c.JSON(200, gin.H{
		"Filtered Logs": logs, //return logs
	})

}
