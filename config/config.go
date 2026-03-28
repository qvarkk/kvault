package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `default:"false"`

	Api    ApiConfig
	DB     DBConfig
	Redis  RedisConfig
	Worker WorkerConfig
}

type ApiConfig struct {
	Port int `default:"8080"`
}

type DBConfig struct {
	Host     string `required:"true"`
	Port     int    `default:"5432"`
	Database string `required:"true"`
	Username string `required:"true"`
	Password string `required:"true"`
}

type RedisConfig struct {
	Host     string `default:"localhost"`
	Port     int    `default:"6379"`
	User     string `required:"true"`
	Password string `required:"true"`
}

type WorkerConfig struct {
	ConcurrentTasks int `default:"10"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
