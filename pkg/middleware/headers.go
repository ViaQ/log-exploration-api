package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func AddHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Header("Access-Control-Allow-Origin", "*")
		gctx.Header("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
		gctx.Header("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		if gctx.Request.Method == "OPTIONS" {
			gctx.AbortWithStatus(204)
		}
		gctx.Next()
	}
}

func TokenHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		tokenValue := gctx.Request.Header["Authorization"]
		if len(tokenValue) == 0 || len(strings.Split(tokenValue[0], "Bearer ")) <= 1 {
			gctx.AbortWithStatusJSON(401, map[string][]string{"Unauthorized, Please pass the token": {"authorization token not found"}})
		}
		gctx.Next()
	}
}
