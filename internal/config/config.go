package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type (
	Config struct {
		Postgres PostgresConfig
		HTTP     HTTPConfig
	}

	PostgresConfig struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
	}

	HTTPConfig struct {
		Host string
		Port string
	}
)

func MustLoad(folder string) (*Config, error) {
	var cfg Config

	configPath := folder + "config.yaml"

	//check if file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
