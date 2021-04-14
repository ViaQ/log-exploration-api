package e2e

import (
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
	repository := esRepository
	params := logs.Parameters{
		Index:      "app",
		Level:      "info",
		FinishTime: "2021-03-17T14:23:20+05:30",
		StartTime:  "2021-03-17T14:22:20+05:30",
		Podname:    "openshift-kube-scheduler-ip-10-0-157-165.ec2.internal",
		Namespace:  "openshift-kube-scheduler",
	}
	logList, err := repository.FilterLogs(params)
	if err != nil {
		t.Error("Error: ", err)
	}
	if logList == nil {
		t.Error("Invalid parameters")
	}
	if !((strings.Contains(logList[0], "index") &&
		strings.Contains(logList[0], "timestamp") &&
		strings.Contains(logList[0], "namespace_name") &&
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
