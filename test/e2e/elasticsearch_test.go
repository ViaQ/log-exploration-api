package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)


var esRepository logs.LogsProvider

func TestMain(m *testing.M) {
	log, _ := initCustomZapLogger("info")

	appConf := configuration.ParseArgs()
	esRepository,_ = elastic.NewElasticRepository(log.Named("elasticsearch"), appConf.Elasticsearch)
	os.Exit(m.Run())
}

func TestFilterPodLogs(t *testing.T) {

	tests := []struct {
		TestName     string
		ShouldFail   bool
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter Pod Logs",
			false,
			map[string]string{"Namespace":"openshift-kube-scheduler","Podname":"openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			nil,
			[]string{"openshift-kube-scheduler","openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
		},
		{
			"Filter Pod logs in a given time range",
			false,
			map[string]string{"StartTime": "2021-03-18T06:41:51.83503Z", "FinishTime": "2021-03-18T06:41:51.83503Z","Podname":"openshift-kube-scheduler-ip-10-0-157-165.ec2.internal","Namespace":"openshift-kube-scheduler"},
			nil,
			[]string{"openshift-kube-scheduler","timestamp", "2021-03-18T06:41:51"},
		},
		{
			"Filter Pod logs by Logging level",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"Podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal",
				"level": "unknown",
			},
			nil,
			[]string{"openshift-kube-scheduler","infra-000001", "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "unknown"},
		},
		{
			"Invalid Podname, or Podname for which no logs exist",
			false,
			map[string]string{
				"Podname":   "hello",
			},
			nil,
			[]string{},
		},
		{
			"Invalid timestamp",
			false,
			map[string]string{
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval for a particular pod",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"Podname":"openshift-kube-scheduler-ip-10-0-157-165.ec2.internal",
				"StartTime":  "2022-03-17T14:22:20+05:30",
				"FinishTime": "2022-03-17T14:23:20+05:30",
			},
			nil,
			[]string{},
		},
		{
			"Negative Limit value",
			false,
			map[string]string{
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}

		addParams(&params,tt.TestParams)

		logList, err := repository.FilterPodLogs(params)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
					}
				}
			}
		}
	}
}

func TestFilterLogsByLabel(t *testing.T) {

	tests := []struct {
		TestName     string
		ShouldFail   bool
		LabelList []string
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter Logs for a set of labels",
			false,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true"},
			map[string]string{},
			nil,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true"},
		},
		{
			"Filter Logs for a set of labels in a given time range",
			false,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true"},
			map[string]string{"StartTime": "2021-03-18T06:41:51.83503Z", "FinishTime": "2021-03-18T06:41:51.83503Z"},
			nil,
			[]string{"timestamp", "2021-03-18T06:41:51.835","app=openshift-kube-scheduler","revision=8","scheduler=true"},
		},
		{
			"Filter logs by Labels, and Logging level",
			false,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true"},
			map[string]string{
				"level": "unknown",
			},
			nil,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true","unknown"},
		},
		{
			"Invalid Labels, or Labels for which no logs exist",
			false,
			[]string{"app=dummy"},
			map[string]string{
			},
			nil,
			[]string{},
		},
		{
			"Invalid timestamp",
			false,
			[]string{"app=openshift-kube-scheduler","revision=8"},
			map[string]string{
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval for a set of labels",
			false,
			[]string{"app=openshift-kube-scheduler","revision=8","scheduler=true"},
			map[string]string{
				"StartTime":  "2022-03-17T14:22:20Z",
				"FinishTime": "2022-03-17T14:23:20Z",
			},
			nil,
			[]string{},
		},
		{
			"Negative Limit value",
			false,
			[]string{},
			map[string]string{
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}

		addParams(&params,tt.TestParams)

		logList, err := repository.FilterLabelLogs(params,tt.LabelList)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
					}
				}
			}
		}
	}
}

func TestFilterNamespaceLogs(t *testing.T) {
	tests := []struct {
		TestName     string
		ShouldFail   bool
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter by namespace",
			false,
			map[string]string{"Namespace":"openshift-kube-scheduler"},
			nil,
			[]string{"openshift-kube-scheduler"},
		},
		{
			"Filter logs of a Namespace, between a given time range",
			false,
			map[string]string{
				"Namespace":  "openshift-kube-scheduler",
				"StartTime":  "2021-03-18T06:41:51.83503Z",
				"FinishTime": "2021-03-18T06:41:51.83503Z",
			},
			nil,
			[]string{"infra-000001", "openshift-kube-scheduler", "2021-03-18T06:41:51.835"},
		},
		{
			"Invalid parameters - Invalid Namespace",
			false,
			map[string]string{
				"Namespace": "world",
			},
			nil,
			[]string{},
		},
		{
			"Invalid timestamp",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval for a namespace",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"StartTime":  "2022-03-17T14:22:20Z",
				"FinishTime": "2022-03-17T14:23:20Z",
			},
			nil,
			[]string{},
		},
		{
			"Negative maxlogs - Incorrect Limit Parameter",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}
		addParams(&params,tt.TestParams)

		logList, err := repository.FilterNamespaceLogs(params)

		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
					}
				}
			}
		}
	}
}

func TestFilterContainerLogs(t *testing.T) {
	tests := []struct {
		TestName     string
		ShouldFail   bool
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter, and Fetch logs for a container in a pod for a given namespace",
			false,
			map[string]string{"ContainerName":"kube-scheduler-cert-syncer","Namespace":"openshift-kube-scheduler","Podname":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal"},
			nil,
			[]string{"kube-scheduler-cert-syncer","openshift-kube-scheduler"},
		},
		{
			"Filter container logs in a given time range",
			false,
			map[string]string{
				"ContainerName":    "kube-scheduler-cert-syncer",
				"Podname":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal",
				"Namespace":  "openshift-kube-scheduler",
				"StartTime":  "2021-03-18T06:41:15.54171Z",
				"FinishTime": "2021-03-18T06:41:18.54171Z",
			},
			nil,
			[]string{"openshift-kube-scheduler", "kube-scheduler-cert-syncer", "2021-03-18T06:41"},
		},
		{
			"Invalid namespace and container",
			false,
			map[string]string{
				"ContainerName":   "hello",
				"Namespace": "world",
				"Podname":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal",
			},
			nil,
			[]string{},
		},
		{
			"Invalid timestamp",
			false,
			map[string]string{
				"Namespace":"openshift-kube-scheduler",
				"Podname":"openshift-kube-scheduler-ip-10-0-162-9.ec2.internal",
				"ContainerName":"kube-scheduler-cert-syncer",
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval",
			false,
			map[string]string{
				"Namespace":"openshift-image-registry",
				"Podname":"image-registry-78b76b488f-bzgqz",
				"ContainerName":"registry",
				"StartTime":  "2022-03-17T14:22:20Z",
				"FinishTime": "2022-03-17T14:23:20Z",
			},
			nil,
			[]string{},
		},
		{
			"Negative maxlogs",
			false,
			map[string]string{
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}
		addParams(&params,tt.TestParams)
		logList, err := repository.FilterContainerLogs(params)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found: %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found: %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found: %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
						break
					}
				}
			}
		}
	}

}

func TestFilterLogs(t *testing.T) {
	tests := []struct {
		TestName     string
		ShouldFail   bool
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter by no parameters",
			false,
			map[string]string{},
			nil,
			[]string{"index", "_id", "_source", "level"},
		},
		{
			"Filter by index",
			false,
			map[string]string{"Index": "infra"},
			nil,
			[]string{"index", "timestamp", "infra-000001"},
		},
		{
			"Filter by time",
			false,
			map[string]string{"StartTime": "2021-03-18T06:41:51.83503Z", "FinishTime": "2021-03-18T06:41:51.83503Z"},
			nil,
			[]string{"timestamp", "2021-03-18T06:41:51"},
		},
		{
			"Filter by podname",
			false,
			map[string]string{"Podname": "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal"},
			nil,
			[]string{"openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "2021-03-18T06:41:51"},
		},
		{
			"Filter by multiple parameters",
			false,
			map[string]string{
				"Index":      "infra",
				"Podname":    "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal",
				"Namespace":  "openshift-kube-scheduler",
				"StartTime":  "2021-03-18T06:41:51.83503Z",
				"FinishTime": "2021-03-18T06:41:51.83503Z",
			},
			nil,
			[]string{"infra-000001", "openshift-kube-scheduler", "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal", "2021-03-18T06:41:51"},
		},
		{
			"Invalid parameters",
			false,
			map[string]string{
				"Podname":   "hello",
				"Namespace": "world",
			},
			nil,
			[]string{},
		},
		{
			"Invalid timestamp",
			false,
			map[string]string{
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval",
			false,
			map[string]string{
				"StartTime":  "2022-03-17T14:22:20Z",
				"FinishTime": "2022-03-17T14:23:20Z",
			},
			nil,
			[]string{},
		},
		{
			"Negative maxlogs",
			false,
			map[string]string{
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}
		addParams(&params,tt.TestParams)
		logList, err := repository.FilterLogs(params)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
					}
				}
			}
		}
	}
}

func TestLogs(t *testing.T) {
	tests := []struct {
		TestName     string
		ShouldFail   bool
		TestParams   map[string]string
		TestError    error
		TestKeywords []string
	}{
		{
			"Filter by no parameters",
			false,
			map[string]string{},
			nil,
			[]string{"index", "_id", "_source", "level"},
		},
		{
			"Filter by time",
			false,
			map[string]string{"StartTime": "2021-03-18T06:41:51.83503Z", "FinishTime": "2021-03-18T06:41:51.83503Z"},
			nil,
			[]string{"timestamp", "2021-03-18T06:41:51"},
		},
		{
			"Filter by Logging level, limitting number of logs to 10",
			false,
			map[string]string{
				"Level":"unknown",
				"Maxlogs":"10",
			},
			nil,
			[]string{"unknown"},
		},
		{
			"Invalid timestamp",
			false,
			map[string]string{
				"StartTime":  "hey",
				"FinishTime": "hey",
			},
			logs.InvalidTimeStamp(),
			[]string{},
		},
		{
			"No logs in the given time interval",
			false,
			map[string]string{
				"StartTime":  "2022-03-17T14:22:20Z",
				"FinishTime": "2022-03-17T14:23:20Z",
			},
			nil,
			[]string{},
		},
		{
			"Negative maxlogs",
			false,
			map[string]string{
				"Maxlogs":  "-2",
			},
			logs.InvalidLimit(),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}
		addParams(&params,tt.TestParams)
		logList, err := repository.FilterLogs(params)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is: %v, found %v", tt.TestError, err)
		} else if logList != nil {
			if strings.Contains(tt.TestName, "Invalid") || strings.Contains(tt.TestName, "No logs") {
				if !strings.Contains(logList[0], "No logs are present or the entry does not exist") {
					t.Errorf("Expected response: No logs are present or the entry does not exist")
				}
			} else {
				for _, keyword := range tt.TestKeywords {
					if !strings.Contains(logList[0], keyword) {
						t.Errorf("Invalid logs found!")
					}
				}
			}
		}
	}
}
func addParams(params *logs.Parameters, testParams map[string]string) {
	for k, v := range testParams {
		switch k {
		case "ContainerName":
			params.ContainerName = v
		case "Namespace":
			params.Namespace = v
		case "Index":
			params.Index = v
		case "Podname":
			params.Podname = v
		case "StartTime":
			params.StartTime = v
		case "FinishTime":
			params.FinishTime = v
		case "Level":
			params.Level = v
		case "Maxlogs":
			params.MaxLogs = v
		}
	}
}


func initCustomZapLogger(level string) (*zap.Logger, error) {
	lv := zap.AtomicLevel{}
	err := lv.UnmarshalText([]byte(strings.ToLower(level)))
	if err != nil {
		return nil, err
	}

	cfg := zap.Config{
		Level:             lv,
		OutputPaths:       []string{"stdout"},
		DisableCaller:     false,
		DisableStacktrace: false,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "msg",
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalLevelEncoder,
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}
	return cfg.Build()
}
