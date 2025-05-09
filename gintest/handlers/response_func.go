package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func successResp(ctx *gin.Context, msg ...any) {
	ctx.JSON(http.StatusOK, msg)
}
func serverErrorResp(ctx *gin.Context, msg ...any) {
	ctx.JSON(http.StatusInternalServerError, msg)
}
func ErrorResp(ctx *gin.Context, msg ...any) {
	ctx.JSON(http.StatusInternalServerError, msg)
}
