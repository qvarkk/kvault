package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `default:"true"`

	Api    ApiConfig
	DB     DBConfig
	Redis  RedisConfig
	Cache  CacheConfig
	Aws    AwsConfig
	Worker WorkerConfig
}

type ApiConfig struct {
	Port int `default:"8080"`
}

type DBConfig struct {
	Host     string `default:"localhost"`
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
	QueueDb  int    `default:"0" envconfig:"QUEUE_DB"`
	CacheDb  int    `default:"1" envconfig:"CACHE_DB"`
}

type CacheConfig struct {
	Enabled      bool          `default:"true"`
	ItemsTtl     time.Duration `default:"5m"`
	FilesTtl     time.Duration `default:"5m"`
	TagsTtl      time.Duration `default:"15m"`
	StopwordsTtl time.Duration `default:"30m"`
}

type AwsConfig struct {
	AccessKeyID     string        `required:"true" envconfig:"ACCESS_KEY_ID"`
	SecretAccessKey string        `required:"true" envconfig:"SECRET_ACCESS_KEY"`
	Region          string        `required:"true" envconfig:"REGION"`
	EndpointUrl     string        `required:"true" envconfig:"ENDPOINT_URL"`
	S3Bucket        string        `required:"true" envconfig:"S3_BUCKET"`
	UrlExpiration   time.Duration `default:"60s"   envconfig:"URL_EXPIRATION"`
	PublicEndpointUrl string      `envconfig:"PUBLIC_ENDPOINT_URL"`
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
