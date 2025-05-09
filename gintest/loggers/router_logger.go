package loggers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func RouterLogger() {
	log.Println("Router Logger initialization:")
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("ReqMethod: %-4v ReqPath: %-20v Handler: %-70v nuHandlers: %+2v", httpMethod, absolutePath, handlerName, nuHandlers)
	}
}
