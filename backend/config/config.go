package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Debug bool `default:"true"`

	Api     ApiConfig
	DB      DBConfig
	Redis   RedisConfig
	Cache   CacheConfig
	Storage StorageConfig
	Worker  WorkerConfig
	Auth    AuthConfig
}

type ApiConfig struct {
	Port        int      `default:"8080"`
	CorsOrigins []string `default:"http://localhost:5173" envconfig:"CORS_ORIGINS"`
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

type StorageConfig struct {
	UploadDir       string `default:"/var/lib/kvault/uploads" envconfig:"UPLOAD_DIR"`
	MaxUploadSizeMB int64  `default:"512" envconfig:"STORAGE_MAX_UPLOAD_SIZE_MB"`
}

type WorkerConfig struct {
	ConcurrentTasks int           `default:"10"`
	MaxRetries      int           `default:"3" envconfig:"MAX_RETRIES"`
	RetryTimeout    time.Duration `default:"5m" envconfig:"RETRY_TIMEOUT"`
}

type AuthConfig struct {
	ApiKeyTtl time.Duration `default:"720h" envconfig:"API_KEY_TTL"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../.env")
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
