package redis

import (
	"errors"

	"github.com/hibiken/asynq"
)

type Redis struct {
	Client *asynq.Client
}

type Config struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedis(config Config) (*Redis, error) {
	redisConnOpt := asynq.RedisClientOpt{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	}

	client := asynq.NewClient(redisConnOpt)

	if err := client.Ping(); err != nil {
		cErr := client.Close()
		return nil, errors.Join(err, cErr)
	}

	return &Redis{
		Client: client,
	}, nil
}
