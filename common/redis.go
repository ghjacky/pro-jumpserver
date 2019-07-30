package common

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type redisConfig struct {
	host     string
	port     int
	password string
	db       int
}

var Redis *redis.Client

func initRedis() {
	Log.Info("Connecting to redis .......")
	Redis = redis.NewClient(&redis.Options{
		Network:      "tcp",
		Addr:         fmt.Sprintf("%s:%d", Config.redisConfig.host, Config.redisConfig.port),
		Password:     Config.redisConfig.password,
		DB:           Config.redisConfig.db,
		ReadTimeout:  3 * time.Second,
		DialTimeout:  5 * time.Second,
		MaxRetries:   3,
		MinIdleConns: 3,
	})
	if _, err := Redis.Ping().Result(); err != nil {
		Log.Fatalf("Couldn't connect to redis at %s:%d", Config.redisConfig.host, Config.redisConfig.port)
	} else {
		Log.Infof("Connected to redis at %s:%d successfully !", Config.redisConfig.host, Config.redisConfig.port)
	}
}
