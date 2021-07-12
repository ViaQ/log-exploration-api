package readiness

import (
	"flag"
	"github.com/ViaQ/log-exploration-api/pkg/configuration"
	"github.com/ViaQ/log-exploration-api/pkg/elastic"
	"github.com/ViaQ/log-exploration-api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewReadinessController(router *gin.Engine) {
	router.GET("/ready", ReadinessHandler)
}

func ReadinessHandler(gctx *gin.Context) {
	middleware.AddHeader()
	esAddress := flag.Lookup("es-addr").Value.(flag.Getter).Get().(string)
	esTLS := flag.Lookup("es-tls").Value.(flag.Getter).Get().(bool)
	esCert := flag.Lookup("es-cert").Value.(flag.Getter).Get().(string)
	esKey := flag.Lookup("es-key").Value.(flag.Getter).Get().(string)
	elasticConfig := configuration.ElasticsearchConfig{
		EsAddress: esAddress,
		EsCert:    esCert,
		EsKey:     esKey,
		UseTLS:    esTLS,
	}
	_, err := elastic.CreateElasticConfig(&elasticConfig)
	if err != nil {
		gctx.JSON(http.StatusOK, gin.H{"Message": "failed to connect to esClient"})
		return
	}
	gctx.JSON(http.StatusOK, gin.H{"Message": "Success"})
}
