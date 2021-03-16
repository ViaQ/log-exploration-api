package configuration

type ElasticsearchConfig struct {
	EsAddress string
	EsCert    string
	EsKey     string
	UseTLS    bool
}
