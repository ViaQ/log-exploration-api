package elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

type ElasticRepository struct {
	esClient *elasticsearch.Client
	log      *zap.Logger
}

func NewElasticRepository(log *zap.Logger, config *configuration.ElasticsearchConfig) (logs.LogsProvider, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.EsAddress,
		},
	}

	if config.UseTLS {
		cert, err := tls.LoadX509KeyPair(config.EsCert, config.EsKey)
		if err != nil {
			log.Error("an error occurred while configuring cert", zap.Error(err))
			return nil, err
		}
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			},
		}
	}

	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Error("failed to configure Elasticsearch", zap.Error(err))
		return nil, err
	}

	repository := &ElasticRepository{
		log:      log,
		esClient: esClient,
	}
	return repository, nil
}

func (repository *ElasticRepository) FilterLogs(params logs.Parameters) ([]string, error) {
	numParameters := 0
	var queryBuilder []map[string]interface{}

	if len(params.Namespace) > 0 {
		term := map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.namespace_name": params.Namespace},
		}
		queryBuilder = append(queryBuilder, term)
		numParameters = numParameters + 1
	}

	if len(params.Podname) > 0 {

		term := map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.pod_name": params.Podname},
		}
		queryBuilder = append(queryBuilder, term)
		numParameters = numParameters + 1
	}
	if len(params.Index) > 0 {
		term := map[string]interface{}{
			"term": map[string]interface{}{
				"_index": params.Index},
		}
		queryBuilder = append(queryBuilder, term)
		numParameters = numParameters + 1
	}
	if len(params.StartTime) > 0 && len(params.FinishTime) > 0 {

		timeSubquery := map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"gte": params.StartTime,
					"lte": params.FinishTime,
				},
			},
		}
		queryBuilder = append(queryBuilder, timeSubquery)
		numParameters = numParameters + 1
	}
	if len(params.Level) > 0 {
		term := map[string]interface{}{
			"term": map[string]interface{}{
				"level": params.Level},
		}
		queryBuilder = append(queryBuilder, term)
		numParameters = numParameters + 1
	}

	maxEntries := 1000 //default value in case params.MaxLogs is nil

	if len(params.MaxLogs) > 0 {
		maxLogs, err := strconv.Atoi(params.MaxLogs)
		if err == nil {
			maxEntries = maxLogs
		}
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               queryBuilder,
				"minimum_should_match": numParameters}},
		"size": maxEntries,
	}

	logsList, err := getLogsList(query, repository.esClient, repository.log)

	if err != nil {
		return nil, err
	}

	return logsList, nil
}

func getLogsList(query map[string]interface{}, esClient *elasticsearch.Client, log *zap.Logger) ([]string, error) {

	jsonQuery, err := json.Marshal(query)

	if err != nil {
		log.Error("An error occurred while processing the query", zap.Error(err))
		return nil, err
	}

	var logsList []string // create a slice of type string to append logs to

	var b strings.Builder

	b.WriteString(string(jsonQuery))
	body := strings.NewReader(b.String())

	searchResult, err := esClient.Search(
		esClient.Search.WithContext(context.Background()),
		esClient.Search.WithBody(body),
		esClient.Search.WithTrackTotalHits(true),
		esClient.Search.WithPretty(),
	)

	if err != nil {
		log.Error("failed exec ES query", zap.Error(err))
		return logsList, getError(err)
	}

	var result map[string]interface{}

	err = json.NewDecoder(searchResult.Body).Decode(&result) //convert searchresult to map[string]interface{}
	if err != nil {
		log.Error("Error occurred while decoding JSON", zap.Error(err))
		return logsList, err
	}

	if _, ok := result["hits"]; !ok {
		log.Error("An error occurred while fetching logs", zap.Any("result", result))
		return logsList, logs.NotFoundError()
	}
	logsList = getRelevantLogs(result)

	return logsList, nil
}

func getRelevantLogs(result map[string]interface{}) []string {
	// iterate through the logs and add them to a slice
	var logsList []string
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		log, _ := json.Marshal(hit) //to return logs in JSON

		logsList = append(logsList, string(log))
	}

	if len(logsList) == 0 {
		logsList = append(logsList, "No logs are present or the entry does not exist")
	}

	return logsList
}

func getError(err error) error {
	fmt.Println("An error occurred while getting a response: ", err)
	err = errors.New("An error occurred while fetching logs")
	return err
}
