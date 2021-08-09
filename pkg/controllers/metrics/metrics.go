package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"strconv"
)

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "custom_metric_http_requests_total",
			Help: "Number of get requests.",
		}, []string{"path"},
	)

	responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "custom_metric_response_status",
			Help: "Status of HTTP response",
		},
		[]string{"status"},
	)

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "custom_metric_http_response_time_seconds",
		Help:    "Duration of HTTP requests.",
		Buckets: prometheus.LinearBuckets(0.001, 0.003, 10),
	}, []string{"path"})

	metricList = []prometheus.Collector{
		totalRequests,
		responseStatus,
		httpDuration,
	}
)

func NewMetricsController(log *zap.Logger, router *gin.Engine) {
	RegisterCustomMetrics(log)
	router.Use(MiddlewareMetrics())
	router.GET("/metrics", HandlerMetrics())
}

func RegisterCustomMetrics(log *zap.Logger) {
	for _, metric := range metricList {
		err := prometheus.Register(metric)
		if err != nil {
			log.Error("An error occurred when registering custom metrics with Prometheus.", zap.Error(err))
			return
		}
	}
}

func MiddlewareMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// BEFORE RESPONSE
		path := c.Request.URL.Path
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))

		c.Next()

		// AFTER RESPONSE
		status := strconv.Itoa(c.Writer.Status())
		responseStatus.WithLabelValues(status).Inc()
		totalRequests.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	}
}

func HandlerMetrics() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
