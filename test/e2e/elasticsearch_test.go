package e2e

import (
	"strings"
	"testing"

	"time"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/constants"
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

func TestGetAllLogs(t *testing.T) {
	repository := esRepository
	logList, err := repository.GetAllLogs()
	if err != nil {
		t.Error("Error: ", err)
	}
	if logList == nil {
		t.Error("Failed to fetch the logs!")
	}
	if !((strings.Contains(logList[0], "index") &&
		 strings.Contains(logList[0], "container_name") && 
		 strings.Contains(logList[0], "pod")) ||
	     strings.Contains(logList[0], "No logs")) {
		t.Error("Logs not found!")
	}
}

func TestFilterByIndex(t *testing.T) {
	repository := esRepository
	logList, err := repository.FilterByIndex(constants.InfraIndexName)
	if err != nil {
		t.Error("Error: ", err)
	}
	if logList == nil {
		t.Error("Index not found!")
	}
	if !(strings.Contains(logList[0], "index") || strings.Contains(logList[0], "No logs")) {
		t.Error("Logs not found!")
	}
}

func TestFilterByTime(t *testing.T) {
	repository := esRepository
	t2 := time.Now()
	count := 10
	t1 := t2.Add(time.Duration(-count) * time.Minute)
	logList, err := repository.FilterByTime(t1, t2)
	if err != nil || logList == nil {
		t.Error("Error: ", err)
	}
	if !((strings.Contains(logList[0], "timestamp") &&
		 strings.Contains(logList[0], "pod")) ||
	     strings.Contains(logList[0], "No logs")) {
		t.Error("Logs not found!")
	}
}

func TestFilterByPodname(t *testing.T) {
	repository := esRepository
	logList, err := repository.FilterByPodName("kube-apiserver-ip-10-0-146-1.ec2.internal")
	if err != nil {
		t.Error("Error: ", err)
	}
	if logList == nil {
		t.Error("Invalid podname!")
	}
	if !(strings.Contains(logList[0], "pod_name") ||
	     strings.Contains(logList[0], "No logs")) {
		t.Error("Logs not found!")
	}
}

func TestFilterLogsMultipleParameters(t *testing.T) {
	repository := esRepository
	t2 := time.Now()
	count := 10
	t1 := t2.Add(time.Duration(-count) * time.Minute)
	logList, err := repository.FilterLogsMultipleParameters("kube-apiserver-ip-10-0-146-1.ec2.internal", "openshift-kube-apiserver", t1, t2)
	if err != nil {
		t.Error("Error: ", err)
	}
	if logList == nil {
		t.Error("Invalid podname/namepsace!")
	}
	if !((strings.Contains(logList[0], "namespace_name") &&
		strings.Contains(logList[0], "pod_name")) ||
	     strings.Contains(logList[0], "No logs")) {
		t.Error("Logs not found!")
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
