package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"gintest/handlers"
	"gintest/handlers/validators"
	"gintest/loggers"
	"gintest/mydb"
	"gintest/routers"
	my_utils "gintest/utils"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var router *gin.Engine = gin.Default()

// 自定义HTTP配置
var server *http.Server = &http.Server{
	Addr:           ":8888",
	Handler:        router,
	ReadTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	MaxHeaderBytes: 1 << 20,
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	// 生成私钥和公钥并写入到yaml配置文件中
	my_utils.InitKey()
	mydb.InitAllModel()
	// router := gin.New()
	// 验证器
	validate := validator.New() // 创建新的验证器实例
	validators.RegisterCustomValiadtor(validate)

	// 自定义日志格式
	loggers.LoggerFormat(router)
	// loggers.RouterLogger()

	// 注册全局中间件
	router.Use(gin.Recovery())                     //  自动恢复
	router.Use(handlers.CheckLoginStatusHandler()) // 验证登录状态

	// Session 设置
	// 生成key
	randomData := make([]byte, 32)
	rand.Read(randomData)
	hash := sha256.Sum256(randomData)
	// 设置cookie，gin使用session插件
	store := cookie.NewStore(hash[:])
	router.Use(sessions.Sessions("mysessions", store))

	// 调用路径Router设置对应Handler
	routers.SetUpUserRoutes(router)
	routers.SetUpCookieRoutes(router, validate)
	routers.SetUpAuthRoutes(router)

	// Server设置
	// 协程启动Server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	// 延时关闭
	// 阻塞，直到接收到Interrupt Signal
	<-ctx.Done()
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")
	// The context is used to inform the server will shutdown after 5 second
	// And the request it is currently handing
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown err:", err)
	}
	log.Println("Server exiting now!")
}
