package elastic

import (
	"errors"
	"strings"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/logs"
)

type MockedElasticsearchProvider struct {
	App   map[time.Time][]string
	Infra map[time.Time][]string
	Audit map[time.Time][]string
}

func NewMockedElastisearchProvider() *MockedElasticsearchProvider {
	return &MockedElasticsearchProvider{
		App:   map[time.Time][]string{},
		Infra: map[time.Time][]string{},
		Audit: map[time.Time][]string{},
	}
}

func (m *MockedElasticsearchProvider) PutDataIntoIndex(index string, data []string) error {
	switch strings.ToLower(index) {
	case "app":
		m.App[time.Now()] = data
		return nil
	case "infra":
		m.Infra[time.Now()] = data
		return nil
	case "audit":
		m.Infra[time.Now()] = data
		return nil
	default:
		return errors.New("unknown index")
	}
}

func (m *MockedElasticsearchProvider) FilterByIndex(index string) ([]string, error) {
	var lg map[time.Time][]string
	switch strings.ToLower(index) {
	case "app":
		lg = m.App
	case "infra":
		lg = m.Infra
	case "audit":
		lg = m.Audit
	}
	if len(lg) == 0 {
		return nil, logs.NotFoundError()
	}

	result := []string{}
	for _, v := range lg {
		result = append(result, v...)
	}
	return result, nil
}

func (m *MockedElasticsearchProvider) FilterByTime(startTime time.Time, finishTime time.Time) ([]string, error) {
	return nil, nil
}

func (m *MockedElasticsearchProvider) GetAllLogs() ([]string, error) {
	return nil, nil
}

func (m *MockedElasticsearchProvider) FilterByPodName(podName string) ([]string, error) {
	return nil, nil
}

func (m *MockedElasticsearchProvider) FilterLogsMultipleParameters(podName string, namespace string, startTime time.Time, finishTime time.Time) ([]string, error) {
	return nil, nil
}
