package Config

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"reflect"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
)

func setElasticSearchParams() (*string, *string) {
	//esAddress := flag.String("es-address", "https://localhost:9200", "ElasticSearch Server Address")
	esCert := flag.String("es-cert", "admin-cert", "admin-cert file location")
	fmt.Println(reflect.TypeOf(esCert))
	fmt.Println("jh")
	esKey := flag.String("es-key", "admin-key", "admin-key file location")
	return esCert, esKey
}

func InitializeElasticSearchClient() *elasticsearch.Client {
	esCert, esKey := setElasticSearchParams()
	flag.Parse() //fetch command line parameters
	var esClient *elasticsearch.Client

	cert, err := tls.LoadX509KeyPair(*esCert, *esKey)
	if err != nil {
		fmt.Println(err)
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://localhost:9200",
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert}},
		},
	}
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("No Error", esClient)
	}

	return esClient

}

func AddESClientToContext(esClient *elasticsearch.Client) gin.HandlerFunc {
	//add ESClient to context
	return func(c *gin.Context) {
		c.Set("esClient", esClient)
		c.Next()
	}
}
