package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	//"reflect"
	"strings"
	"net/http"
    "github.com/gin-gonic/gin"
	"flag"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func main() {

	esAddress := flag.String("es-address", "https://localhost:9200", "ElasticSearch Server Address")
	esCert := flag.String("es-cert","admin-cert","admin-cert file location")
	esKey := flag.String("es-key","admin-key","admin-key file location")
	flag.Parse() //fetch command line parameters
	var esClient *elasticsearch.Client

	esClient = InitializeElasticSearchClient(esAddress,esCert,esKey,esClient)

	r := gin.Default() //initialise

	r.Use(AddESClientToContext(esClient))
	//	endpoint to get all logs
	r.GET("/", getAllLogs)

	//endpoint to get infrastructure logs
	r.GET("/infra", getInfrastructureLogs)

	//endpoint to get application logs
	r.GET("/app", getApplicationLogs)

	//endpoint to get audit logs
	r.GET("/audit", getAuditLogs)

	//endpoint to filter logs by start and finish time - please enter time in the following format- HH:MM:SS
	r.GET("/filter/:startTime/:finishTime", filterByTime)

	r.Run() //run server on port 8080
}
func filterByTime(c *gin.Context){

	startTime:= c.Params.ByName("startTime")
	finishTime:= c.Params.ByName("finishTime")

	fmt.Println(startTime)
	fmt.Println(finishTime)

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch esClient from context
	if !ok {
		fmt.Println("An Error Occurred")
	}

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

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func InitializeElasticSearchClient(esAddress *string,esCert *string,esKey *string,esClient *elasticsearch.Client) *elasticsearch.Client {


	cert, err := tls.LoadX509KeyPair(*esCert, *esKey)
	if err != nil {
		fmt.Println(err)
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert}},
		},
	}
	esClient, err = elasticsearch.NewClient(cfg)
	if(err!=nil) {
		fmt.Println("Error", err)
	}else{
		fmt.Println("No Error",esClient)
	}

	return esClient


}
func getInfrastructureLogs(c *gin.Context) {

	esClient,ok := c.MustGet("esClient").(*elasticsearch.Client) //fetch ESClient from context
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

	if(err!=nil) {
		fmt.Println(err)
	}

	var logs[] string

	 logs = getRelevantLogs(searchResult) // create a slice of type string to append logs to

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getAllLogs(c *gin.Context){
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

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getApplicationLogs(c *gin.Context){

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
	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})
}
func getAuditLogs(c *gin.Context){

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

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func AddESClientToContext(esClient *elasticsearch.Client) gin.HandlerFunc {
	//add ESClient to context
	return func(c *gin.Context) {
		c.Set("esClient", esClient)
		c.Next()
	}
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


