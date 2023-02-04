package config

import "github.com/spf13/viper"

type ServerConfigSet struct {
	Port  int               `mapstructure:"port"`
	Users map[string]string `mapstructure:"users"`
}

type ClientConfigSet struct {
	Uid        string
	RemotePort uint16
	LocalPort  uint16
	Host       string
}

type ConfigSet interface{}

func ReadConfigFile(path, format string) (ConfigSet, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	switch v.GetString("mode") {
	case "client":
		return &ClientConfigSet{
			Uid:        v.GetString("uuid"),
			RemotePort: v.GetUint16("rPort"),
			LocalPort:  v.GetUint16("lPort"),
			Host:       v.GetString("host"),
		}, nil
	default:
		return &ServerConfigSet{
			Port:  v.GetInt("port"),
			Users: v.GetStringMapString("users"),
		}, nil
	}
}
