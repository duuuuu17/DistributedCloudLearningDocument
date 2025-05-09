package handlers

import (
	"fmt"
	"gintest/mydb"
	"gintest/mydb/my_redis"
	"gintest/types"
	my_utils "gintest/utils"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func PostHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := &types.PostReq{}
		err := ctx.ShouldBindJSON(req)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Invaild JSON data",
			})
			ctx.Abort()
			return
		}
		ctx.JSON(http.StatusOK, req)
	}
}

// 表单使用shouldBind方法绑定
func FormHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var myForm types.MyForm
		ctx.ShouldBind(&myForm)
		ctx.JSON(http.StatusOK, gin.H{
			"color": myForm.Colors,
		})
	}
}
func UserFormHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var myForm types.User
		ctx.ShouldBindJSON(&myForm)
		ctx.JSON(http.StatusOK, myForm)
	}
}

// 绑定uri
func UserIDFormHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var myForm types.UserInfo
		ctx.ShouldBindUri(&myForm)
		ctx.JSON(http.StatusOK, myForm)
	}
}

// 设置cookie
func SetCookieHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SetCookie("username", "ethan", 3600, "/", "localhost", false, true)
		ctx.JSON(http.StatusOK, gin.H{
			"message": "success set cookie!",
		})
	}
}

// 获取cookie
func GETCookieHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("username") // 读取cookie
		if err != nil {
			ctx.JSON(http.StatusNonAuthoritativeInfo, gin.H{
				"message": "cookie expired or not set!",
			})
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": cookie,
		})
	}
}

// 使用内置验证器
func ValidationPasswordHandler(validate *validator.Validate) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var user types.TestUser
		//获取当前请求的Session
		session := sessions.Default(ctx)
		session.Set("username", "john_doe") // 设置session键值对
		session.Save()                      // 保存session更改
		// 绑定JSON
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		// 手动验证结构体
		if err := validate.Struct(user); err != nil {
			// 获取验证失败的字段错误信息
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				var errorMessages []string
				for _, validationError := range validationErrors {
					errorMessages = append(errorMessages, fmt.Sprintf("Field %s: %s", validationError.Field(), validationError.Error()))
				}
				ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMessages})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "validation define!"})
			}
			return
		}
		ctx.SecureJSON(http.StatusOK, gin.H{"success": user})
	}
}
func DownloadFileHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := http.Get("https://raw.githubusercontent.com/gin-gonic/logo/master/color.png")
		if err != nil && resp.StatusCode != http.StatusOK {
			ctx.Status(http.StatusServiceUnavailable)
		}
		// 文件reader
		reader := resp.Body
		defer reader.Close()
		// 内容长度
		contentLength := resp.ContentLength
		// 内容类型
		contentType := resp.Header.Get("Content-Type")
		// 设置回复头
		extraHeadrs := map[string]string{
			"Content-Disposition": `attachment; filename="gopher.png"`,
		}
		// 使用DataFromReader读取[]byte数据
		ctx.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeadrs)
	}
}
func GetHello1Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SecureJSON(http.StatusOK, gin.H{
			"message": "hello Gin!",
		})
	}
}

// 设置认证handler，不过一般不建议使用gin.Basic，无状态使用token
func PostAuthorizationHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := types.LoginUser{}
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"msg": "Request data format failure!",
				"err": err,
			})
			return
		}
		token := my_utils.GenerateJWT(&user)
		ctx.SecureJSON(http.StatusOK, gin.H{
			"token": token,
		})
	}
}

// 绑定User
func getUserFromContext(ctx *gin.Context, user *types.LoginUser) {
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "Request data format failure!",
			"err": err,
		})
	}
}

// 注册用户
func RegisterUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := &types.LoginUser{}
		getUserFromContext(ctx, user)
		if ok := mydb.RegisterUserIntoDB(user); !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "register failure wait minutes!",
			})
			return
		}
		token := my_utils.GenerateJWT(user)
		ctx.SecureJSON(http.StatusOK, gin.H{
			"message": "success registered!",
			"token":   token,
		})
	}
}

// 验证token
func validationUserToken(ctx *gin.Context, token *types.Token) bool {
	if token.Token != "" {
		rdbToken, _ := my_redis.RDB.Get(ctx, token.Username).Result() // 获取token
		// 如果redis存在token，进行解析是否超时
		if my_utils.EqualToken(token.Token, rdbToken) && !my_utils.ParseTokenCheckExpired(token.Token) {
			return true
		}
	}
	return false
}

// 检查登录状态,即验证token是否过期
func CheckLoginStatusHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 全局处理器，跳过指定路径
		excludePath := []string{"/user/login"}
		for _, path := range excludePath {
			if path == ctx.Request.URL.Path {
				ctx.Next()
				return
			}
		}
		token := &types.Token{}
		ctx.ShouldBindBodyWithJSON(token)
		// 从redis string获取保存的token
		if !validationUserToken(ctx, token) {
			successResp(ctx, map[string]interface{}{
				"statusCode": "200",
				"msg":        "token have been invalidated, try to login!",
			})
			return
		}
		ctx.Next()
		successResp(ctx, map[string]interface{}{
			"message": "success login check by redis!",
		})

	}
}

// 登录用户
func LoginUserHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user := &types.LoginUser{}
		getUserFromContext(ctx, user)
		// redis不存在token时，需要检查登录密码，以及生成token并返回，还有缓存到redis中
		if ok := mydb.QueryUserInfo(user); !ok {
			serverErrorResp(ctx, map[string]interface{}{
				"error": "Input User or Password value was failure, or the account not register!",
			})
			return
		}
		// 生成token
		token := my_utils.GenerateJWT(user)
		my_redis.RDB.Set(ctx, user.Username, token, 24*time.Hour)
		successResp(ctx, map[string]interface{}{
			"message": "success login & generator token!",
			"token":   token,
		})
	}
}
