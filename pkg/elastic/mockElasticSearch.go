package elastic

import (
	"errors"
	"strings"
	"time"

	"github.com/ViaQ/log-exploration-api/pkg/logs"
)

type MockedElasticsearchProvider struct {
	App            map[time.Time][]string
	Infra          map[time.Time][]string
	Audit          map[time.Time][]string
	checkReadiness bool
}

func (m *MockedElasticsearchProvider) CheckReadiness() bool {
	if m.checkReadiness == true {
		return true
	} else {
		return false
	}
}

func NewMockedElastisearchProvider() *MockedElasticsearchProvider {
	return &MockedElasticsearchProvider{
		App:            map[time.Time][]string{},
		Infra:          map[time.Time][]string{},
		Audit:          map[time.Time][]string{},
		checkReadiness: true,
	}
}

const (
	Level         = "level: "
	containerName = "container_name: "
	podName       = "pod_name: "
	namespaceName = "namespace_name: "
)

func (m *MockedElasticsearchProvider) UpdateReadinessState(checkReadiness bool) {
	m.checkReadiness = checkReadiness
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

func mockedFilterHelper(params logs.Parameters, m *MockedElasticsearchProvider) ([]string, error) {
	lg := make(map[time.Time][]string)
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
	var result []string
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
	if len(result) > 0 {
		return result, nil
	} else {
		return nil, logs.NotFoundError()
	}
}

func generateIntermediateLogs(typeOfLog string, parameter string, logs []string) []string {
	var resultantLogs []string
	for _, v := range logs {
		str := typeOfLog + parameter
		if strings.Contains(v, str) {
			resultantLogs = append(resultantLogs, v)
		}
	}
	return resultantLogs
}

func (m *MockedElasticsearchProvider) Logs(params logs.Parameters) ([]string, error) {
	resultantLogs, err := mockedFilterHelper(params, m)
	if err != nil {
		return nil, logs.NotFoundError()
	}
	if len(params.Level) > 0 {
		resultantLogs = generateIntermediateLogs(Level, params.Level, resultantLogs)
	}
	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterLogs(params logs.Parameters) ([]string, error) {

	tempLogsStore, err := mockedFilterHelper(params, m)
	var resultantLogs []string
	if err != nil {
		return nil, logs.NotFoundError()
	}
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
	resultantLogs, err := mockedFilterHelper(params, m)
	if err != nil {
		return nil, logs.NotFoundError()
	}
	resultantLogs = generateIntermediateLogs(namespaceName, params.Namespace, resultantLogs)
	resultantLogs = generateIntermediateLogs(podName, params.Podname, resultantLogs)
	resultantLogs = generateIntermediateLogs(containerName, params.ContainerName, resultantLogs)
	if len(params.Level) > 0 {
		resultantLogs = generateIntermediateLogs(Level, params.Level, resultantLogs)
	}
	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterLabelLogs(params logs.Parameters, labelList []string) ([]string, error) {
	resultantLogs, err := mockedFilterHelper(params, m)
	if err != nil {
		return nil, logs.NotFoundError()
	}
	for _, label := range labelList {
		tempLogsStore := resultantLogs
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
	if len(params.Level) > 0 {
		resultantLogs = generateIntermediateLogs(Level, params.Level, resultantLogs)
	}
	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterPodLogs(params logs.Parameters) ([]string, error) {
	resultantLogs, err := mockedFilterHelper(params, m)
	if err != nil {
		return nil, logs.NotFoundError()
	}
	resultantLogs = generateIntermediateLogs(namespaceName, params.Namespace, resultantLogs)
	resultantLogs = generateIntermediateLogs(podName, params.Podname, resultantLogs)
	if len(params.Level) > 0 {
		resultantLogs = generateIntermediateLogs(Level, params.Level, resultantLogs)
	}
	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}

func (m *MockedElasticsearchProvider) FilterNamespaceLogs(params logs.Parameters) ([]string, error) {
	resultantLogs, err := mockedFilterHelper(params, m)
	if err != nil {
		return nil, logs.NotFoundError()
	}
	resultantLogs = generateIntermediateLogs(namespaceName, params.Namespace, resultantLogs)
	if len(params.Level) > 0 {
		resultantLogs = generateIntermediateLogs(Level, params.Level, resultantLogs)
	}
	if len(resultantLogs) == 0 {
		return nil, logs.NotFoundError()
	}
	return resultantLogs, nil
}
