package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   `yaml:"server"`
	Postgres `yaml:"postgres"`
	Kafka    `yaml:"kafka"`
}

type Server struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

type Postgres struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT"`
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname" env:"POSTGRES_DB"`
}

type Kafka struct {
	Broker  string `yaml:"broker" env:"KAFKA_BROKER"`
	Topic   string `yaml:"topic" env:"KAFKA_TOPIC"`
	GroupID string `yaml:"group_id" env:"KAFKA_GROUP_ID"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
	}
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}
	return &cfg, nil
}
