package loggers

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerFormat(router *gin.Engine) {
	router.Use(gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s\" %s\"\n",
			params.ClientIP,
			params.TimeStamp.Format(time.RFC1123),
			params.Method,
			params.Path,
			params.Request.Proto,
			params.StatusCode,
			params.Latency,
			params.Request.UserAgent(),
			params.ErrorMessage,
		)
	}))
	// 记录到文件
	f, _ := os.Create("gin.log")
	// 同时将日志写入文件和控制台
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
}
