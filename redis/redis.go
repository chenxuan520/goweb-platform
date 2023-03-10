package redis

import (
	"context"

	"github.com/chenxuan520/goweb-platform/logger"
	"github.com/go-redis/redis/v8"
)

//redis连接
var _defaultRedis *redis.Client

func Init(addr, password string, db int) (r *redis.Client, err error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       db,       // use default DB
	})
	_, err = client.Ping(context.Background()).Result()
	_defaultRedis = client

	return
}

func GetRedis() *redis.Client {
	if _defaultRedis == nil {
		logger.GetLogger().Error("redis is not initialized")
		return nil
	}
	return _defaultRedis
}
