package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	PostgresURL string `mapstructure:"postgres_url"`
	VectorDBURL string `mapstructure:"vector_db_url"`
}

func Load() (*Config, error) {
	viper.SetConfigName("database")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Enable env var override: DATABASE_POSTGRES_URL -> database.postgres_url
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
