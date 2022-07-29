package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func successResponse(data interface{}) gin.H {
	return gin.H{
		"code":      "200",
		"message":   "ok",
		"data":      data,
		"timestamp": time.Now().UnixMilli(),
	}
}

func failResponse(code string, message string) gin.H {
	return gin.H{
		"code":      code,
		"message":   message,
		"timestamp": time.Now().UnixMilli(),
	}
}

func BadRequest(ctx *gin.Context, code string, message string) {
	ctx.JSON(http.StatusInternalServerError, failResponse(code, message))
}

func ServerError(ctx *gin.Context, code string, message string) {
	ctx.JSON(http.StatusInternalServerError, failResponse(code, message))
}

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, successResponse(data))
}

func Unauthorized(ctx *gin.Context, code string, message string) {
	ctx.JSON(http.StatusUnauthorized, failResponse(code, message))
}
