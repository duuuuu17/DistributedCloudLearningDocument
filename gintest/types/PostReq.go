package types

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PostReq struct {
	Id   int32  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// 对于表单使用tag form声明绑定字段
type MyForm struct {
	Colors []string `form:"colors[]"`
}

// 当指定时间格式时，需要自定义序列化和反序列化时间格式处理
type CustomTime time.Time

const timeFormat = "2006-01-02 15:04:05"

// 序列化时间字段
func (c CustomTime) MarshalJSON() ([]byte, error) {
	t := time.Time(c)                         // 将自定义类型转化为time.Time
	return json.Marshal(t.Format(timeFormat)) // 通过format进行格式化为字符串
}

// 反序列化时间字段：先将字段使用string接收，通过time.Parse根据自定义时间格式
func (c *CustomTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ts, err := time.Parse(timeFormat, s)
	if err != nil {
		return err
	}
	*c = CustomTime(ts)
	return nil
}

type User struct {
	Name string     `json:"name"`
	Ts   CustomTime `json:"timestamp"`
	Addr string     `json:"address" binding:"required, prefix=Addr-"`
}

// uri绑定
type UserInfo struct {
	ID   string `uri:"id" binding:"required,uuid"`
	Name string `uri:"name" binding:"required"`
}

// 自定义结构体验证器的结构体
type TestUser struct {
	Username        string `json:"username" binding:"required,min=3,max=20"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_passowrd" binding:"required"`
}
type LoginUser struct {
	gorm.Model
	Username string `json:"username" binding:"required,min=3,max=20" gorm:"<-; not null; uniqueIndex"`
	Password string `json:"password" binding:"required,min=6" gorm:"<-; not null;index"`
}
type Token struct {
	Username string `json:"username" binding:"required,min=3,max=20" `
	Token    string `json:"token" gorm:"-"`
}
