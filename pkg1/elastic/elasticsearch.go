package elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"logexplorationapi/pkg1/constants"
	"logexplorationapi/pkg1/logs"
	"net/http"
	"strings"
)

type ElasticRepository struct {
	esClient *elasticsearch.Client
}

func InitializeElasticSearchClient() *elasticsearch.Client {

	esAddress := flag.String("es-address", "https://localhost:9200", "ElasticSearch Server Address")
	esCert := flag.String("es-cert", "admin-cert", "admin-cert file location")
	esKey := flag.String("es-key", "admin-key", "admin-key file location")
	flag.Parse() //fetch command line parameters
	cert, err := tls.LoadX509KeyPair(*esCert, *esKey)
	if err != nil {
		fmt.Println(err)
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			*esAddress,
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert}},
		},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("Error", err)
	}

	return esClient
}

func NewElasticRepository() logs.LogsProvider {
	esClient := InitializeElasticSearchClient()
	repository := &ElasticRepository{esClient: esClient}
	return repository
}

func (repository *ElasticRepository) FilterByIndex(index string) []string {
	esClient := repository.esClient
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(index),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	var logs []string // create a slice of type string to append logs to

	logs = getRelevantLogs(searchResult)

	return logs

}
func (repository *ElasticRepository) FilterByTime(startTime string, finishTime string) []string {

	fmt.Println(startTime)
	fmt.Println(finishTime)

	var logs []string // create a slice of type string to append logs to
	//
	esClient := repository.esClient
	query := fmt.Sprintf(`{
		"query": {
		"range" : {
			"@timestamp" : {
				"gte": "%s",
  				"lte": "%s",
 				"time_zone": "+00:00"
			}
		}
	}
	}`, startTime, finishTime)

	var b strings.Builder
	b.WriteString(query)
	body := strings.NewReader(b.String())
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(constants.InfraIndexName, constants.AuditIndexName, constants.AppIndexName),
		esClient.Search.WithBody(body),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	logs = getRelevantLogs(searchResult)
	return logs
}
func (repository *ElasticRepository) FilterByPodName(podName string) []string {
	//

	var logs []string // create a slice of type string to append logs to

	esClient := repository.esClient

	query := fmt.Sprintf(`{"query": {
					"match" : {
						    "kubernetes.pod_name":{"query":"%s"}
						  }
					}
				}`, podName)
	var b strings.Builder
	b.WriteString(query)
	body := strings.NewReader(b.String())

	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithBody(body),
		esClient.Search.WithIndex(constants.InfraIndexName, constants.AuditIndexName, constants.AppIndexName),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}
	logs = getRelevantLogs(searchResult)

	return logs

}

func (repository *ElasticRepository) GetAllLogs() []string {
	esClient := repository.esClient
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(constants.InfraIndexName, constants.AuditIndexName, constants.AppIndexName),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)
	if err != nil {
		fmt.Println("Error getting response: %s", err)
	}

	var logs []string // create a slice of type string to append logs to

	logs = getRelevantLogs(searchResult)

	return logs

}
func getRelevantLogs(searchResult *esapi.Response) []string {

	var result map[string]interface{}
	error := json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if error != nil {
		fmt.Println("Error occurred while decoding JSON ", error)
	}
	fmt.Println(result)

	var logs []string // create a slice of type string to append logs to
	if _, ok := result["hits"]; !ok {
		logs = append(logs, "No logs found")
		return logs
	}
	// iterate through the logs and add them to a slice

	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log := fmt.Sprintf("%v", hit)
		logs = append(logs, log)
		fmt.Println(hit, "\n")
		fmt.Println()
	}

	if len(logs) == 0 {
		logs = append(logs, "No logs Present")
	}

	return logs
}
