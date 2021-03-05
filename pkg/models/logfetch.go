package models

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func FilterByTime(esClient *elasticsearch.Client, startTime string, finishTime string) []string{

	fmt.Println(startTime)
	fmt.Println(finishTime)

	var result map[string]interface{}
	var logs[] string // create a slice of type string to append logs to
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("infra-000001","app-000001","audit-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}
	json.NewDecoder(searchResult.Body).Decode(&result)
	fmt.Println(result)
	if(err!=nil){
		fmt.Println(err)
	}

	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) { //iterate through logs and check if timestamp lies between start and end times
		log := fmt.Sprintf("%v", hit)
		index := strings.Index(log, "@timestamp")
		fmt.Println(index)
		time := log[index+22 : index+30]
		if (time >= startTime && time <= finishTime) {
			logs = append(logs, log)
			fmt.Println(hit, "\n")
			fmt.Println()
		}
	}

	return logs

}

func GetInfrastructureLogs(esClient *elasticsearch.Client) []string{

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

	if(err!=nil) {
		fmt.Println(err)
	}

	var logs[] string

	logs = getRelevantLogs(searchResult) // create a slice of type string to append logs to


		return logs//return logs


}
func GetAllLogs(esClient *elasticsearch.Client) []string{
	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("app-000001","infra-000001","audit-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	var logs[] string // create a slice of type string to append logs to

	logs = getRelevantLogs(searchResult)

	return logs

}
func GetApplicationLogs(esClient *elasticsearch.Client) []string{

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

	//fetch logs from index app-000001
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex("app-000001"),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	var logs[] string

	logs = getRelevantLogs(searchResult)
	return logs
}
func GetAuditLogs(esClient *elasticsearch.Client) []string{

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
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
	var logs[] string
	logs = getRelevantLogs(searchResult)

	return logs

}

func getRelevantLogs(searchResult *esapi.Response) []string{

	var result map[string]interface{}

	json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	fmt.Println(result)

	var logs[] string // create a slice of type string to append logs to
	//iterate through the logs and add them to a slice
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log:= fmt.Sprintf("%v",hit)
		logs = append(logs,log)
		fmt.Println(hit,"\n")
		fmt.Println()
	}
	return logs
}

