package config

import (
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
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 9002)

	// Bind env vars for docker-compose compatibility
	_ = viper.BindEnv("database.postgres_url", "POSTGRES_URL")
	_ = viper.BindEnv("database.vector_db_url", "VECTOR_DB_URL")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.port", "SERVER_PORT")

	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional when using env vars (e.g., in Docker)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
