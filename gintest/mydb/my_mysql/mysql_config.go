package my_mysql

import (
	my_utils "gintest/utils"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MySQLDB *gorm.DB

func InitDB() {
	var dialector gorm.Dialector = mysql.New(mysql.Config{
		DSN:                       "root:123456@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	})
	var err error
	MySQLDB, err = gorm.Open(dialector, &gorm.Config{})
	my_utils.ErrorInLog(err)
	sqlDB, err := MySQLDB.DB() // return object was type of sqlDB
	my_utils.ErrorInLog(err)
	// 设置连接数和连接最大存活时间
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetMaxOpenConns(40)
	sqlDB.SetMaxIdleConns(20)
	err = sqlDB.Ping()
	if err == nil {
		log.Println("MySQL start succession!")
	}
}
