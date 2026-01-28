package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	AI      AIServiceConfig
	Gateway GatewayServiceConfig
	Redis   RedisConfig
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

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func Load() (*Config, error) {
	viper.SetConfigName("codebase")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("gateway.host", "localhost")
	viper.SetDefault("gateway.port", 8080)
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Enable env var override: REDIS_HOST -> redis.host
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	for _, key := range viper.AllKeys() {
		viper.BindEnv(key)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
