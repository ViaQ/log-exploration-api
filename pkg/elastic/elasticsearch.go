package elastic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/constants"
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

type ElasticRepository struct {
	esClient *elasticsearch.Client
	log      *zap.Logger
}

const (
	Term          = "term"
	Match         = "match"
	NamespaceName = "kubernetes.namespace_name"
	PodName       = "kubernetes.pod_name"
	ContainerName = "kubernetes.container_name.raw"
	FlatLabel     = "kubernetes.flat_labels"
)

func CreateElasticConfig(config *configuration.ElasticsearchConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			config.EsAddress,
		},
	}

	if config.UseTLS {
		cert, err := tls.LoadX509KeyPair(config.EsCert, config.EsKey)
		if err != nil {
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
		return nil, err
	}
	return esClient, nil
}
func NewElasticRepository(log *zap.Logger, config *configuration.ElasticsearchConfig) (logs.LogsProvider, error) {
	esClient, err := CreateElasticConfig(config)
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
func generateLogs(queryBuilder []map[string]interface{}, params logs.Parameters, repository *ElasticRepository) ([]string, error) {
	if len(params.Index) > 0 {
		term := map[string]interface{}{
			"term": map[string]interface{}{
				"_index": params.Index},
		}
		queryBuilder = append(queryBuilder, term)
	}

	if len(params.StartTime) > 0 && len(params.FinishTime) > 0 {
		startTime, _ := time.Parse(time.RFC3339Nano, params.StartTime)

		finishTime, _ := time.Parse(time.RFC3339Nano, params.FinishTime)

		timeSubQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"gte": startTime,
					"lte": finishTime,
				},
			},
		}
		queryBuilder = append(queryBuilder, timeSubQuery)
	}
	if len(params.Level) > 0 {
		term := map[string]interface{}{
			"term": map[string]interface{}{
				"level": params.Level},
		}
		queryBuilder = append(queryBuilder, term)
	}
	maxEntries := 1000
	if len(params.MaxLogs) > 0 {
		maxLogs, _ := strconv.Atoi(params.MaxLogs)

		maxEntries = maxLogs

	}

	var sortQuery []map[string]interface{}
	sortSubQuery := map[string]interface{}{
		"@timestamp": map[string]interface{}{
			"order": "desc"},
	}
	sortQuery = append(sortQuery, sortSubQuery)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": queryBuilder,
			}},
		"size": maxEntries,
		"sort": sortQuery,
	}
	logsList, err := getLogsList(query, repository.esClient, repository.log)
	if err != nil {
		return nil, err
	}
	return logsList, nil
}

func appendToQueryBuilder(key string, typeOfQuery string, value interface{}) map[string]interface{} {
	query := map[string]interface{}{
		typeOfQuery: map[string]interface{}{
			key: value},
	}
	return query
}
func (repository *ElasticRepository) FilterPodLogs(params logs.Parameters) ([]string, error) {
	err := validateParams(params)
	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}
	var queryBuilder []map[string]interface{}
	queryBuilder = append(queryBuilder, appendToQueryBuilder(NamespaceName, Term, params.Namespace))
	queryBuilder = append(queryBuilder, appendToQueryBuilder(PodName, Term, params.Podname))
	return generateLogs(queryBuilder, params, repository)
}

func (repository *ElasticRepository) FilterNamespaceLogs(params logs.Parameters) ([]string, error) {
	err := validateParams(params)
	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}
	var queryBuilder []map[string]interface{}
	queryBuilder = append(queryBuilder, appendToQueryBuilder(NamespaceName, Term, params.Namespace))
	return generateLogs(queryBuilder, params, repository)
}

func (repository *ElasticRepository) FilterLabelLogs(params logs.Parameters, labelsList []string) ([]string, error) {
	err := validateParams(params)
	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}
	var queryBuilder []map[string]interface{}
	for _, label := range labelsList {
		value := map[string]interface{}{
			"query": label, "operator": "AND"}
		queryBuilder = append(queryBuilder, appendToQueryBuilder(FlatLabel, Match, value))
	}
	return generateLogs(queryBuilder, params, repository)
}

func (repository *ElasticRepository) FilterContainerLogs(params logs.Parameters) ([]string, error) {
	err := validateParams(params)
	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}
	var queryBuilder []map[string]interface{}
	queryBuilder = append(queryBuilder, appendToQueryBuilder(NamespaceName, Term, params.Namespace))
	queryBuilder = append(queryBuilder, appendToQueryBuilder(PodName, Term, params.Podname))
	queryBuilder = append(queryBuilder, appendToQueryBuilder(ContainerName, Term, params.ContainerName))
	return generateLogs(queryBuilder, params, repository)
}
func (repository *ElasticRepository) Logs(params logs.Parameters) ([]string, error) {

	err := validateParams(params)
	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}
	var queryBuilder []map[string]interface{}
	return generateLogs(queryBuilder, params, repository)

}

func (repository *ElasticRepository) FilterLogs(params logs.Parameters) ([]string, error) {

	err := validateParams(params)

	if err != nil {
		repository.log.Error("Invalid Query Parameters:", zap.Error(err))
		return nil, err
	}

	var queryBuilder []map[string]interface{}

	if len(params.Namespace) > 0 {
		namespaceSubQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.namespace_name": params.Namespace},
		}
		queryBuilder = append(queryBuilder, namespaceSubQuery)
	}
	if len(params.Podname) > 0 {
		podnameSubQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"kubernetes.pod_name": params.Podname},
		}
		queryBuilder = append(queryBuilder, podnameSubQuery)
	}
	if len(params.Index) > 0 {
		indexSubQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"_index": params.Index},
		}
		queryBuilder = append(queryBuilder, indexSubQuery)
	}
	if len(params.StartTime) > 0 && len(params.FinishTime) > 0 {

		startTime, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finishTime, _ := time.Parse(time.RFC3339Nano, params.FinishTime)

		timeSubquery := map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"gte": startTime,
					"lte": finishTime,
				},
			},
		}
		queryBuilder = append(queryBuilder, timeSubquery)
	}
	if len(params.Level) > 0 {
		levelSubQuery := map[string]interface{}{
			"term": map[string]interface{}{
				"level": params.Level},
		}
		queryBuilder = append(queryBuilder, levelSubQuery)
	}
	maxEntries := 1000 //default value in case params.MaxLogs is nil
	if len(params.MaxLogs) > 0 {
		maxLogs, _ := strconv.Atoi(params.MaxLogs)
		maxEntries = maxLogs
	}

	var sortQuery []map[string]interface{}
	sortSubQuery := map[string]interface{}{
		"@timestamp": map[string]interface{}{
			"order": "desc"},
	}
	sortQuery = append(sortQuery, sortSubQuery)

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": queryBuilder,
			}},
		"size": maxEntries,
		"sort": sortQuery,
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
		esClient.Search.WithIndex(constants.InfraIndexName, constants.AppIndexName, constants.AuditIndexName),
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

	err = errors.New("an error occurred while fetching logs")
	return err
}
