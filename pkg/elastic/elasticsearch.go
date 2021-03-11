package elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"logexplorationapi/pkg/configuration"
	"logexplorationapi/pkg/constants"
	"logexplorationapi/pkg/logs"
	"net/http"
	"strings"
	"time"
)

type ElasticRepository struct {
	esClient *elasticsearch.Client
}

func NewElasticRepository(config *configuration.ApplicationConfiguration) (logs.LogsProvider, error) {

	cert, err := tls.LoadX509KeyPair(config.EsCert, config.EsKey)
	if err != nil {
		fmt.Println("An error occurred while configuring cert- ", err)
		err = errors.New("An error occurred while configuring cert. Please check if es-cert and es-key are valid")
		return nil, err
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.EsAddress,
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert}},
		},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("Fialed to configure ElasticSearch- ", err)
		err = errors.New("Fialed to configure ElasticSearch")
		return nil, err
	}

	repository := &ElasticRepository{esClient: esClient}
	return repository, nil
}

func (repository *ElasticRepository) FilterByIndex(index string) ([]string, error) {

	esClient := repository.esClient
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(index),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)

	var logsList []string // create a slice of type string to append logs to

	if err != nil {
		return logsList, getError(err)
	}

	var result map[string]interface{}

	err = json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if err != nil {
		fmt.Println("Error occurred while decoding JSON ", err)
		err = errors.New("An Internal error occurred")
		return logsList, err
	}

	if _, ok := result["hits"]; !ok {
		fmt.Println("An error occurred while fetching logs - ", result)
		return logsList, logs.NotFoundError()
	}

	logsList = getRelevantLogs(result)

	return logsList, nil

}
func (repository *ElasticRepository) FilterByTime(startTime time.Time, finishTime time.Time) ([]string, error) {

	var logsList []string // create a slice of type string to append logs to

	//splitting date-time and timezone to populate the query
	start := strings.Split(startTime.String(), " ")[0] //format- YYYY-MM-DDTHH:MM:SS
	finish := strings.Split(finishTime.String(), " ")[0]
	timezone := strings.Split(startTime.String(), " ")[2] //format- +0000
	esClient := repository.esClient
	query := fmt.Sprintf(`{
		"query": {
		"range" : {
			"@timestamp" : {
				"gte": "%s",
  				"lte": "%s",
				"time_zone":"%s"
			}
		}
	}
	}`, start, finish, timezone)

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
		return logsList, getError(err)
	}
	var result map[string]interface{}
	err = json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if err != nil {
		fmt.Println("Error occurred while decoding JSON ", err)
		err = errors.New("An Internal error occurred")
		return logsList, err
	}

	if _, ok := result["hits"]; !ok {
		fmt.Println("An error occurred while fetching logs..Result obtained is null", result)

		return logsList, logs.NotFoundError()
	}

	logsList = getRelevantLogs(result)

	return logsList, nil
}
func (repository *ElasticRepository) FilterByPodName(podName string) ([]string, error) {

	var logsList []string // create a slice of type string to append logs to

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
		return logsList, getError(err)
	}
	var result map[string]interface{}
	err = json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if err != nil {
		fmt.Println("Error occurred while decoding JSON ", err)
		err = errors.New("An Internal error occurred")
		return logsList, err
	}

	if _, ok := result["hits"]; !ok {
		fmt.Println("An error occurred while fetching logs..Result obtained is null", result)

		return logsList, logs.NotFoundError()
	}

	logsList = getRelevantLogs(result)

	return logsList, nil

}

func (repository *ElasticRepository) GetAllLogs() ([]string, error) {
	esClient := repository.esClient
	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithIndex(constants.InfraIndexName, constants.AuditIndexName, constants.AppIndexName),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)

	var logsList []string // create a slice of type string to append logs to

	if err != nil {
		return logsList, getError(err)
	}
	var result map[string]interface{}
	err = json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if err != nil {
		fmt.Println("Error occurred while decoding JSON ", err)
		err = errors.New("An Internal error occurred")
		return logsList, err
	}

	if _, ok := result["hits"]; !ok {
		fmt.Println("An error occurred while fetching logs..Result obtained is null", result)
		err := errors.New("An Error Occurred..")
		return logsList, err
	}

	logsList = getRelevantLogs(result)

	return logsList, nil
}
func getRelevantLogs(result map[string]interface{}) []string {

	// iterate through the logs and add them to a slice
	var logsList []string
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log := fmt.Sprintf("%v", hit)
		logsList = append(logsList, log)
	}

	if len(logsList) == 0 {
		logsList = append(logsList, "No logs Present or the entry does not exist")
	}

	return logsList
}

func getError(err error) error {

	fmt.Println("An Error occurred while getting a response: ", err)
	err = errors.New("An Error occurred while fetching logs")
	return err

}
