package config

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	HttpPort    int    `json:"http_port" toml:"http_port"`
	HttpHost    string `json:"http_host" toml:"http_host"`
	MetricsPort int    `json:"metrics_port" toml:"metrics_port"`
}

var _cfg *Config = nil

func (conf *Config) MetricPort() int {
	return conf.MetricsPort
}

func ParseConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Error("get config failed", "err", err)
		panic(err)
	}
	err = toml.Unmarshal(data, &_cfg)
	// err = json.Unmarshal(data, &_cfg)
	if err != nil {
		log.Error("unmarshal config failed", "err", err)
		panic(err)
	}
	return _cfg, nil
}

func GetConfig() *Config {
	return _cfg
}

var (
	DefaultCors    = []string{"localhost"} // Default cors domain for the apis
	DefaultVhosts  = []string{"localhost"} // Default virtual hosts for the apis
	DefaultOrigins = []string{"localhost"} // Default origins for the apis
	DefaultPrefix  = ""                    // Default prefix for the apis
	DefaultModules = []string{"time", "block"}
)

const (
	APIBatchItemLimit         = 2000
	APIBatchResponseSizeLimit = 250 * 1000 * 1000
)
