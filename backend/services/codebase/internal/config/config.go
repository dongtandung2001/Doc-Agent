package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	AI      AIServiceConfig
	Gateway GatewayServiceConfig
}

type AIServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type GatewayServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	viper.SetConfigName("codebase")
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
