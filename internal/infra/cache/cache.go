package cache

import (
	"github.com/go-redis/redis/v8"
	"gx1727.com/xin/pkg/config"
)

var Client *redis.Client

func Init(cfg *config.RedisConfig) {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func Get() *redis.Client {
	return Client
}
