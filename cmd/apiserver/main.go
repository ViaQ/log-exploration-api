package main

import (
	healthcontroller "github.com/ViaQ/log-exploration-api/pkg/controllers/health"
	logscontroller "github.com/ViaQ/log-exploration-api/pkg/controllers/logs"
	metricscontroller "github.com/ViaQ/log-exploration-api/pkg/controllers/metrics"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/ViaQ/log-exploration-api/pkg/version"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"

	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	appConf := configuration.ParseArgs()

	log, err := initCustomZapLogger(appConf.LogLevel)
	if err != nil {
		panic(err)
	}

	log.Info("application started", zap.Any("configuration", appConf),
		zap.String("version", version.Version),
		zap.String("build_time", version.BuildTime))

	repository, err := elastic.NewElasticRepository(log.Named("elasticsearch"), appConf.Elasticsearch)
	if err != nil {
		log.Error("unable to create elasticsearch repo", zap.Error(err))
		return
	}

	router := gin.New()
	metricscontroller.NewMetricsController(log.Named("metrics"), router)
	logscontroller.NewLogsController(log.Named("logs-controller"), repository, router)
	healthcontroller.NewHealthController(router, repository)

	router.Run()
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
