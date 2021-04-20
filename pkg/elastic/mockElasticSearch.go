package elastic

import (
	"errors"
	"fmt"
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

func (m *MockedElasticsearchProvider) FilterLogs(params logs.Parameters) ([]string, error) {
	lg := make(map[time.Time][]string)
	fmt.Print("Params: ", params.Index, params.Namespace, params.Podname, params.FinishTime)
	if len(params.Index) > 0 {
		switch strings.ToLower(params.Index) {
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

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
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
	if len(params.Podname) > 0 {
		for _, v := range temp {
			pod := "pod_name: " + params.Podname
			if strings.Contains(v, pod) {
				result = append(result, v)
			}
		}
		temp = result
		result = []string{}
	}
	if len(params.Namespace) > 0 {
		for _, v := range temp {
			ns := "namespace_name: " + params.Namespace
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
