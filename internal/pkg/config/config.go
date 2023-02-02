package config

import "github.com/spf13/viper"

type ServerConfigSet struct {
	Port  int               `mapstructure:"port"`
	Users map[string]string `mapstructure:"users"`
}

func ReadConfigFile(path, format string) (*ServerConfigSet, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return &ServerConfigSet{
		Port:  v.GetInt("port"),
		Users: v.GetStringMapString("users"),
	}, nil
}
