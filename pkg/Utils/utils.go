package Utils

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func GetRelevantLogs(searchResult *esapi.Response) []string {

	var result map[string]interface{}

	json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	fmt.Println(result)

	var logs []string // create a slice of type string to append logs to
	//iterate through the logs and add them to a slice
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log := fmt.Sprintf("%v", hit)
		logs = append(logs, log)
		fmt.Println(hit, "\n")
		fmt.Println()
	}
	return logs
}
