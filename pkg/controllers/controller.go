package controllers

import (
	//"context"
	//"encoding/json"
	"fmt"
	"logexplorationapi/pkg/models"
	//"reflect"
	//"strings"
	"github.com/gin-gonic/gin"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func FilterByTime(c *gin.Context){

	startTime:= c.Params.ByName("startTime")
	finishTime:= c.Params.ByName("finishTime")

	fmt.Println(startTime)
	fmt.Println(finishTime)

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}


	var logs[] string // create a slice of type string to append logs to

	logs = models.FilterByTime(esClient,startTime,finishTime)
	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})
}

func GetInfrastructureLogs(c *gin.Context) {

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	var logs[] string

	//fetch logs from index infra-000001
	logs = models.GetInfrastructureLogs(esClient) // create a slice of type string to append logs to

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func GetAllLogs(c *gin.Context){
	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	var logs[] string // create a slice of type string to append logs to

	logs = models.GetAllLogs(esClient)

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func GetApplicationLogs(c *gin.Context){

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	var logs[] string

	logs =models.GetApplicationLogs(esClient)
	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})
}
func GetAuditLogs(c *gin.Context){

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	var logs[] string
	logs = models.GetAuditLogs(esClient)

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}


