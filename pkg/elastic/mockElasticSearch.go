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
		m.Audit[time.Now()] = data
		return nil
	default:
		return errors.New("unknown index")
	}
}

func (m *MockedElasticsearchProvider) PutDataAtTime(logTime time.Time, index string, data []string) error {
	switch strings.ToLower(index) {
	case "app":
		m.App[logTime] = data
		return nil
	case "infra":
		m.Infra[logTime] = data
		return nil
	case "audit":
		m.Audit[logTime] = data
		return nil
	default:
		return errors.New("unknown index")
	}
}

func (m *MockedElasticsearchProvider) FilterLogs(index string, podname string, namespace string,
	starttime string, finishtime string, level string, maxlogs string) ([]string, error) {
	lg := make(map[time.Time][]string)
	if len(index) > 0 {
		switch strings.ToLower(index) {
		case "app":
			lg = m.App
		case "infra":
			lg = m.Infra
		case "audit":
			lg = m.Audit
		}
	} else {
		for k, v := range m.App {
			if v != nil {
				lg[k] = v
			}
		}
		for k, v := range m.Infra {
			if v != nil {
				lg[k] = v
			}
		}
		for k, v := range m.Audit {
			if v != nil {
				lg[k] = v
			}
		}
	}

	if len(lg) == 0 {
		return nil, logs.NotFoundError()
	}

	result := []string{}
	temp := []string{}

	if len(finishtime) > 0 && len(starttime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, starttime)
		finish, _ := time.Parse(time.RFC3339Nano, finishtime)
		for k, v := range lg {
			if k.After(start) && k.Before(finish) {
				result = append(result, v...)
			}
		}
	} else {
		for _, v := range lg {
			result = append(result, v...)
		}
	}

	temp = result
	result = []string{}
	if len(podname) > 0 {
		for _, v := range temp {
			pod := "pod_name: " + podname
			if strings.Contains(v, pod) {
				result = append(result, v)
			}
		}
		temp = result
		result = []string{}
	}
	if len(namespace) > 0 {
		for _, v := range temp {
			ns := "namespace_name: " + namespace
			if strings.Contains(v, ns) {
				result = append(result, v)
			}
		}
	} else {
		result = temp
	}

	if len(result) == 0 {
		return nil, logs.NotFoundError()
	}
	return result, nil
}
