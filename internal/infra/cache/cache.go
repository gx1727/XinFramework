package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/xin-framework/xin/configs"
)

var Client *redis.Client

func Init(cfg *configs.RedisConfig) {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func Get() *redis.Client {
	return Client
}
