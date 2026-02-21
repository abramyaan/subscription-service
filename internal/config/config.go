package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Host         string        `env:"SERVER_HOST" env-default:"0.0.0.0"`
	Port         string        `env:"SERVER_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" env-default:"10s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" env-default:"10s"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-required:"true"` // Ошибка, если не задан
	Password string `env:"DB_PASSWORD" env-required:"true"`
	DBName   string `env:"DB_NAME" env-default:"subscriptions_db"`
	SSLMode  string `env:"DB_SSLMODE" env-default:"disable"`
}

type LogConfig struct {
	Level string `env:"LOG_LEVEL" env-default:"info"`
}

func Load() (*Config, error) {
	var cfg Config

	
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
	}

	return &cfg, nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}