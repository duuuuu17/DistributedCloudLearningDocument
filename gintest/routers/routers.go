package routers

import (
	"gintest/handlers"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SetUpUserRoutes(r *gin.Engine) {
	u := r.Group("/user")
	u.POST("/ping", handlers.PostHandler())
	u.POST("/color", handlers.FormHandler())
	u.POST("/info", handlers.UserFormHandler())
	// 声明uri路由处理时，通过":声明字段"进行绑定
	u.GET("/:name/:id", handlers.UserIDFormHandler())
	// use gorm query
	u.POST("/register", handlers.RegisterUserHandler())
	u.POST("/login", handlers.LoginUserHandler())
}
func SetUpCookieRoutes(e *gin.Engine, validate *validator.Validate) {
	e.GET("/set_cookie", handlers.SetCookieHandler())
	e.GET("/get_cookie", handlers.GETCookieHandler())
	e.GET("/hello", handlers.DownloadFileHandler())
	e.POST("/hello1", handlers.GetHello1Handler())
}

// 设置认证路由,一般使用无状态token:jwt进行验证
func SetUpAuthRoutes(e *gin.Engine) {
	users := gin.Accounts{
		"admin": "admin",
		"user":  "123456",
	}
	{
		authorized := e.Group("/admin", gin.BasicAuth(users))
		authorized.POST("/dashboard", handlers.PostAuthorizationHandler())
	}
}
