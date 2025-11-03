package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Backend BackendConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type BackendConfig struct {
	CodebaseService ServiceEndpoint `mapstructure:"codebase"`
	DocgenService   ServiceEndpoint `mapstructure:"docgen"`
	AIService       ServiceEndpoint `mapstructure:"ai"`
}

type ServiceEndpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetConfigName("gateway")
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
