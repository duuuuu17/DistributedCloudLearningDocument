package mydb

import (
	my_mysql "gintest/mydb/my_mysql"
	"gintest/mydb/my_redis"
)

func InitAllModel() {
	my_mysql.InitDB()
	my_redis.InitRedis()
	InitUser()
}
