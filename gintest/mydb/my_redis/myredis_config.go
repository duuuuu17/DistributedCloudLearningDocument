package my_redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	if RDB.Ping(context.Background()).Val() == "PONG" {
		log.Println("Redis start succession!")
	}
}
