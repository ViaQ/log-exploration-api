package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/ViaQ/log-exploration-api/pkg/logs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var esRepository logs.LogsProvider

func TestMain(t *testing.T) {
	log, _ := initCustomZapLogger("info")

	appConf := configuration.ParseArgs()
	esRepository, _ = elastic.NewElasticRepository(log.Named("elasticsearch"), appConf.Elasticsearch)
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
			errors.New("incorrect time format: Please Enter Start Time in the following format YYYY-MM-DDTHH:MM:SS[TIMEZONE ex:+00:00]"),
			[]string{},
		},
		{
			"No logs in the given time interval",
			false,
			map[string]string{
				"StartTime":  "2022-03-17T14:22:20+05:30",
				"FinishTime": "2022-03-17T14:23:20+05:30",
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
			errors.New("invalid max logs limit value, please enter a valid integer from 0 to 1000"),
			[]string{},
		},
	}

	for _, tt := range tests {
		t.Log("Running:", tt.TestName)
		repository := esRepository
		params := logs.Parameters{}
		for k, v := range tt.TestParams {
			switch k {
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
		logList, err := repository.FilterLogs(params)
		if err == nil && tt.TestError != nil {
			t.Errorf("Expected error is %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError == nil {
			t.Errorf("Expected error is %v, found %v", tt.TestError, err)
		} else if err != nil && tt.TestError != nil && err.Error() != tt.TestError.Error() {
			t.Errorf("Expected error is %v, found %v", tt.TestError, err)
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
