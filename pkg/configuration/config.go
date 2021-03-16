package configuration

import "flag"

type ApplicationConfiguration struct {
	LogLevel      string
	Elasticsearch *ElasticsearchConfig
}

func NewApplicationConfiguration() *ApplicationConfiguration {
	return &ApplicationConfiguration{
		Elasticsearch: &ElasticsearchConfig{},
	}
}

func ParseArgs() *ApplicationConfiguration {
	c := NewApplicationConfiguration()

	flag.StringVar(&c.LogLevel, "log-level", "info", "application log level (debug | info | warn | error)")
	flag.BoolVar(&c.Elasticsearch.UseTLS, "es-tls", false, "use TLS for Elasticseach connection")
	flag.StringVar(&c.Elasticsearch.EsAddress, "es-addr", "http://localhost:9200", "Elasticsearch Server Address")
	flag.StringVar(&c.Elasticsearch.EsCert, "es-cert", "admin-cert", "admin-cert file location")
	flag.StringVar(&c.Elasticsearch.EsKey, "es-key", "admin-key", "admin-key file location")
	flag.Parse()

	return c
}
