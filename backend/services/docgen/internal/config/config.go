package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	AI      ServiceEndpoint
	Gateway ServiceEndpoint
}

type ServiceEndpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetConfigName("docgen")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
