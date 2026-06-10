package redis

import "time"

type ConnConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

type CacheConfig struct {
	IsEnabled bool

	ItemsTtl     time.Duration
	FilesTtl     time.Duration
	TagsTtl      time.Duration
	StopwordsTtl time.Duration
}
