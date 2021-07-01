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

func (m *MockedElasticsearchProvider) Cleanup() {
	m.App = map[time.Time][]string{}
	m.Infra = map[time.Time][]string{}
	m.Audit = map[time.Time][]string{}
}

func (m *MockedElasticsearchProvider) Logs(params logs.Parameters) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	for k, v := range m.App {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Infra {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Audit {
		if v != nil {
			logsLists[k] = v
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	if len(params.Level) > 0 {

		tempLogsStore = resultantLogs
		resultantLogs = []string{}

		for _, v := range tempLogsStore {
			level := "level: " + params.Level
			if strings.Contains(v, level) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	}

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterLogs(params logs.Parameters) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	if len(params.Index) > 0 {
		switch strings.ToLower(params.Index) {
		case "app":
			logsLists = m.App
		case "infra":
			logsLists = m.Infra
		case "audit":
			logsLists = m.Audit
		}
	} else {
		for k, v := range m.App {
			if v != nil {
				logsLists[k] = v
			}
		}
		for k, v := range m.Infra {
			if v != nil {
				logsLists[k] = v
			}
		}
		for k, v := range m.Audit {
			if v != nil {
				logsLists[k] = v
			}
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}
	if len(params.Podname) > 0 {
		for _, v := range tempLogsStore {
			pod := "pod_name: " + params.Podname
			if strings.Contains(v, pod) {
				resultantLogs = append(resultantLogs, v)
			}
		}
		tempLogsStore = resultantLogs
		resultantLogs = []string{}
	}
	if len(params.Namespace) > 0 {
		for _, v := range tempLogsStore {
			ns := "namespace_name: " + params.Namespace
			if strings.Contains(v, ns) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	} else {
		resultantLogs = tempLogsStore
	}

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}
func (m *MockedElasticsearchProvider) FilterContainerLogs(params logs.Parameters) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	for k, v := range m.App {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Infra {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Audit {
		if v != nil {
			logsLists[k] = v
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	for _, v := range tempLogsStore {
		ns := "namespace_name: " + params.Namespace
		if strings.Contains(v, ns) {
			resultantLogs = append(resultantLogs, v)
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	for _, v := range tempLogsStore {
		containerName := "container_name: " + params.ContainerName
		if strings.Contains(v, containerName) {
			resultantLogs = append(resultantLogs, v)
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	for _, v := range tempLogsStore {
		podName := "pod_name: " + params.Podname
		if strings.Contains(v, podName) {
			resultantLogs = append(resultantLogs, v)
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	if len(params.Level) > 0 {
		for _, v := range tempLogsStore {
			level := "level: " + params.Level
			if strings.Contains(v, level) {
				resultantLogs = append(resultantLogs, v)
			}
		}
		tempLogsStore = resultantLogs
	}

	resultantLogs = tempLogsStore

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterLabelLogs(params logs.Parameters, labelList []string) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	for k, v := range m.App {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Infra {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Audit {
		if v != nil {
			logsLists[k] = v
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	if len(params.Level) > 0 {

		tempLogsStore = resultantLogs
		resultantLogs = []string{}

		for _, v := range tempLogsStore {
			level := "level: " + params.Level
			if strings.Contains(v, level) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	}

	for _, label := range labelList {
		tempLogsStore = resultantLogs
		resultantLogs = []string{}
		for _, v := range tempLogsStore {
			if label == "" {
				resultantLogs = append(resultantLogs, v)
				continue
			}
			if strings.Contains(v, label) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	}

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterPodLogs(params logs.Parameters) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	for k, v := range m.App {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Infra {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Audit {
		if v != nil {
			logsLists[k] = v
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	if len(params.Level) > 0 {

		tempLogsStore = resultantLogs
		resultantLogs = []string{}

		for _, v := range tempLogsStore {
			level := "level: " + params.Level
			if strings.Contains(v, level) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	for _, v := range tempLogsStore {
		namespace := "namespace_name: " + params.Namespace
		if strings.Contains(v, namespace) {
			resultantLogs = append(resultantLogs, v)
		}
	}
	tempLogsStore = resultantLogs
	resultantLogs = []string{}

	for _, v := range tempLogsStore {
		pod := "pod_name: " + params.Podname
		if strings.Contains(v, pod) {
			resultantLogs = append(resultantLogs, v)
		}
	}

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterNamespaceLogs(params logs.Parameters) ([]string, error) {

	logsLists := make(map[time.Time][]string)

	for k, v := range m.App {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Infra {
		if v != nil {
			logsLists[k] = v
		}
	}
	for k, v := range m.Audit {
		if v != nil {
			logsLists[k] = v
		}
	}

	if len(logsLists) == 0 {
		return nil, logs.NotFoundError()
	}

	resultantLogs := []string{}
	var tempLogsStore []string

	if len(params.FinishTime) > 0 && len(params.StartTime) > 0 {
		start, _ := time.Parse(time.RFC3339Nano, params.StartTime)
		finish, _ := time.Parse(time.RFC3339Nano, params.FinishTime)
		for k, v := range logsLists {
			if k.After(start) && k.Before(finish) {
				resultantLogs = append(resultantLogs, v...)
			}
		}
	} else {
		for _, v := range logsLists {
			resultantLogs = append(resultantLogs, v...)
		}
	}

	if len(params.Level) > 0 {

		tempLogsStore = resultantLogs
		resultantLogs = []string{}

		for _, v := range tempLogsStore {
			level := "level: " + params.Level
			if strings.Contains(v, level) {
				resultantLogs = append(resultantLogs, v)
			}
		}
	}

	tempLogsStore = resultantLogs
	resultantLogs = []string{}
	for _, v := range tempLogsStore {
		namespace := "namespace_name: " + params.Namespace
		if strings.Contains(v, namespace) {
			resultantLogs = append(resultantLogs, v)
		}
	}

	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}
