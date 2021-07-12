package middleware

import "github.com/gin-gonic/gin"

func AddHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Header("Access-Control-Allow-Origin", "*")
		gctx.Next()
	}
}
