package validators

import (
	"gintest/types"
	"strings"

	"github.com/go-playground/validator/v10"
)

// 字段级验证器
// 一般形式：参数validator.FieldLevel
func ValidateUUID(fl validator.FieldLevel) bool {
	_, ok := fl.Field().Interface().(string)
	return ok
}

// 自定义验证函数：验证整数是否为质数
func ValidatePrime(fl validator.FieldLevel) bool {
	num, ok := fl.Field().Interface().(int)
	if !ok || num < 2 {
		return false
	}
	for i := 2; i*i <= num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}

// closure闭包形式：返回验证函数validator.Func
func ValidatorPrefix(prefix string) validator.Func {
	return func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return strings.HasPrefix(value, prefix)
	}
}
func ValidatePassword(sl validator.StructLevel) {
	user := sl.Current().Interface().(types.TestUser)
	if user.Password != user.ConfirmPassword {
		sl.ReportError(user.Password, "Password", "password", "passwordmatch", "")
	}
}
