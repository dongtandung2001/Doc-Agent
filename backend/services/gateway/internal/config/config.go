package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Backends BackendConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type BackendConfig struct {
	CodebaseService   ServiceEndpoint `mapstructure:"codebase"`
	DocgenService     ServiceEndpoint `mapstructure:"docgen"`
	AIService         ServiceEndpoint `mapstructure:"ai"`
	LocalAgentService ServiceEndpoint `mapstructure:"local_agent"`
	DatabaseService   ServiceEndpoint `mapstructure:"database"`
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

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Enable env var override: BACKENDS_CODEBASE_HOST -> backends.codebase.host
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
