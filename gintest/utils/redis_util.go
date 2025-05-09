package my_utils

import (
	"log"

	"github.com/go-redis/redis/v8"
)

// var RedisConnector redis.
func IsRedisNotHaveToken(err error) bool {
	if err == redis.Nil {
		return true
	} else {
		log.Fatalln("redis error:", err)
	}
	return false
}
func EqualToken(t1 string, t2 string) bool {
	return t1 == t2
}
