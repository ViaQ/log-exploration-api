package configuration

import "time"

type ElasticsearchConfig struct {
	EsAddress string
	EsCert    string
	EsKey     string
	UseTLS    bool
	Timeout   time.Duration
}
