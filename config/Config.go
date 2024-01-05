package config

import (
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type Config struct {
	Port        int `json:"port" toml:"port"`
	MetricsPort int `json:"metrics_port" toml:"metrics_port"`
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
