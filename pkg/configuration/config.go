package configuration

import "flag"

type ApplicationConfiguration struct {
	EsAddress string
	EsCert    string
	EsKey     string
}

func NewApplicationConfiguration() *ApplicationConfiguration {
	return &ApplicationConfiguration{
		EsAddress: "https://localhost:9200",
		EsCert:    "admin-cert",
		EsKey:     "admin-key",
	}

}

func ParseArgs() *ApplicationConfiguration {
	c := NewApplicationConfiguration()

	flag.StringVar(&c.EsAddress, "es-addr", "https://localhost:9200", "ElasticSearch Server Address")
	flag.StringVar(&c.EsCert, "es-cert", "admin-cert", "admin-cert file location")
	flag.StringVar(&c.EsKey, "es-key", "admin-key", "admin-key file location")
	flag.Parse()

	return c
}
