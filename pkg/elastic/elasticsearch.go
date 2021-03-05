package elastic

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"net/http"
)
func InitializeElasticSearchClient() *elasticsearch.Client {
	esAddress := flag.String("es-address", "https://localhost:9200", "ElasticSearch Server Address")
	esCert := flag.String("es-cert","admin-cert","admin-cert file location")
	esKey := flag.String("es-key","admin-key","admin-key file location")
	flag.Parse() //fetch command line parameters
	cert, err := tls.LoadX509KeyPair(*esCert, *esKey)
	if err != nil {
		fmt.Println(err)
	}
	cfg := elasticsearch.Config{
		Addresses: []string{
			*esAddress,
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert}},
		},
	}
	esClient, err = elasticsearch.NewClient(cfg)
	if(err!=nil) {
		fmt.Println("Error", err)
	}else{
		fmt.Println("No Error",esClient)
	}

	return esClient

}
