package validators

import (
	"fmt"
	"gintest/types"

	"github.com/go-playground/validator/v10"
)

// 注册验证器
func RegisterCustomValiadtor(validate *validator.Validate) {
	// 自定义验证器，在注册时需要指明验证字段tag
	err := validate.RegisterValidation("uuid", ValidateUUID) //注册刚才的单个字段验证器
	PrintErr(err)
	err = validate.RegisterValidation("prime", ValidatePrime)
	PrintErr(err)
	err = validate.RegisterValidation("prefix", ValidatorPrefix("Addr-"))
	PrintErr(err)
	// 添加自定义结构体验证器
	validate.RegisterStructValidation(ValidatePassword, types.TestUser{})
}
func PrintErr(err error) bool {
	if err != nil {
		fmt.Println("Failed to register custom validator", err)
		return false
	}
	return true
}
