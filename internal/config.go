package internal

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ServerConf struct {
	ListenPort int `yaml:"port"`
}

type ClientConf struct {
	ServerPort        int    `yaml:"server_port"`
	HeartBeatInterval int    `yaml:"interval"`
	RemotePort        int    `yaml:"remote_port"`
	RemoteAddr        string `yaml:"remote_addr"`
	LocalPort         int    `yaml:"local_port"`
	LocalAddr         string `yaml:"local_addr"`
	MaxRetryTimes     int    `yaml:"max_retry_times"`
	CTXTimeout        int    `yaml:"context_timeout"`
	ReplyTimeout      int    `yaml:"reply_timeout"`
}

type Config struct {
	Debug  bool       `yaml:"debug"`
	Mode   string     `yaml:"mode"`
	Server ServerConf `yaml:"server"`
	Client ClientConf `yaml:"client"`
}

func ParseConfig(filename string) (interface{}, bool, error) {
	var conf Config
	fileData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, false, err
	}
	err = yaml.Unmarshal(fileData, &conf)
	if err != nil {
		return nil, false, err
	}
	switch conf.Mode {
	case "server":
		return &conf.Server, conf.Debug, nil
	case "client":
		return &conf.Client, conf.Debug, nil
	default:
		panic("Unknown mode type, server or client")
	}
}
