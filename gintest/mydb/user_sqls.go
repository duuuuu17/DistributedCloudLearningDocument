package mydb

import (
	"context"
	mymysql "gintest/mydb/my_mysql"
	"gintest/types"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

func InitUser() {
	mymysql.MySQLDB.AutoMigrate(&types.LoginUser{})
}

// 注册用户
func RegisterUserIntoDB(user *types.LoginUser) bool {
	// 此处使用了Replicas into的方式。但是，从业务角度上来看，注册肯定不能这样做的
	// 先查询用户是否已存在,若重复注册则更新update、password字段
	tx := mymysql.MySQLDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "username"}},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at", "password"}),
	}).Create(&user)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if tx.Error != nil {
		tx.Logger.Error(ctx, tx.Error.Error())
		return false
	}
	var u types.LoginUser
	mymysql.MySQLDB.Select("*").First(&u)
	log.Printf("Register an new username: %s\n", u.Username)
	return true
}

// 查询User信息
func QueryUserInfo(user *types.LoginUser) bool {
	password := types.LoginUser{} // 如果使用var password string类变量初始声明或初始赋值后就无法更改吗，所以到First中使用也无法
	tx := mymysql.MySQLDB.Select("password").Where("username = ?", user.Username).First(&password)
	if tx.Error != nil || password.Password == "" {
		log.Println("password, ", password)
		return false
	}
	if password.Password != user.Password {
		return false
	}
	return true
}
