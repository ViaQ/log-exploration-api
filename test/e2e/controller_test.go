package e2e

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	logscontroller "github.com/ViaQ/log-exploration-api/pkg/controllers/logs"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logsController *logscontroller.LogsController
var router *gin.Engine
var w *httptest.ResponseRecorder

func TestMain(t *testing.T) {
	log, _ := initCustomZapLogger("info")
	appConf := configuration.ParseArgs()
	logProvider, _ := elastic.NewElasticRepository(log.Named("elasticsearch"), appConf.Elasticsearch)
	router = gin.Default()
	logsController = logscontroller.NewLogsController(log, logProvider, router)
	w = httptest.NewRecorder()
}

func TestForGetAllLogs(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logs/", nil)
	status := w.Code
	router.ServeHTTP(w, req)
	p, err := ioutil.ReadAll(w.Body)
	pageOK := err == nil && strings.Index(string(p), "Logs") > 0

	if status == http.StatusBadRequest {
		t.Error("Bad request!")
		t.Fail()
	}
	if status == http.StatusInternalServerError {
		t.Error("Internal Server error!")
	}
	if !(status == http.StatusOK && pageOK) {
		t.Error("Logs not found!")
		t.Fail()
	}
}

func TestFilterLogsByIndex(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logs/indexfilter/infra-000001", nil)
	status := w.Code
	router.ServeHTTP(w, req)
	p, err := ioutil.ReadAll(w.Body)
	pageOK := err == nil && strings.Index(string(p), "Logs") > 0

	if status == http.StatusBadRequest {
		t.Error("Bad request: Invalid index!")
	}
	if status == http.StatusInternalServerError {
		t.Error("Internal Server error!")
	}
	if !(status == http.StatusOK && pageOK) {
		t.Error("Logs not found!")
	}
}

func TestFilterLogsByTime(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logs/timefilter/2021-03-17T14:22:20+03:30/2021-03-17T14:23:20+05:30", nil)
	status := w.Code
	router.ServeHTTP(w, req)
	p, err := ioutil.ReadAll(w.Body)
	pageOK := err == nil && strings.Index(string(p), "Logs") > 0

	if status == http.StatusBadRequest {
		t.Error("Bad request: Invalid index!")
	}
	if status == http.StatusInternalServerError {
		t.Error("Internal Server error!")
	}
	if !(status == http.StatusOK && pageOK) {
		t.Error("Logs not found!")
	}
}

func TestFilterLogsByPodName(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logs/podnamefilter/kube-apiserver-ip-10-0-146-1.ec2.internal", nil)
	status := w.Code
	router.ServeHTTP(w, req)
	p, err := ioutil.ReadAll(w.Body)
	pageOK := err == nil && strings.Index(string(p), "Logs") > 0

	if status == http.StatusBadRequest {
		t.Error("Bad request: Invalid podname!")
	}
	if status == http.StatusInternalServerError {
		t.Error("Internal Server error!")
	}
	if !(status == http.StatusOK && pageOK) {
		t.Error("Logs not found!")
	}
}

func TestFilterMultipleParameters(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logs/multifilter/kube-apiserver-ip-10-0-146-1.ec2.internal/openshift-kube-apiserver/2021-03-17T14:22:20+03:30/2021-03-17T14:23:20+05:30", nil)
	status := w.Code
	router.ServeHTTP(w, req)
	p, err := ioutil.ReadAll(w.Body)
	pageOK := err == nil && strings.Index(string(p), "Logs") > 0

	if status == http.StatusBadRequest {
		t.Error("Bad request: Invalid parameter values/format!")
	}
	if status == http.StatusInternalServerError {
		t.Error("Internal Server error!")
	}
	if !(status == http.StatusOK && pageOK) {
		t.Error("Logs not found!")
	}
}

func initCustomLogger(level string) (*zap.Logger, error) {
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
