package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
)

func AddHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Header("Access-Control-Allow-Origin", "*")
		gctx.Next()
	}
}

func TokenHeader() gin.HandlerFunc {
	return func(gctx *gin.Context) {
		tokenValue := gctx.Request.Header["Authorization"]
		token := map[string]string{"Authorization": tokenValue[0]}
		mJson, err := json.Marshal(token)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		jsonStr := string(mJson)
		err = os.Setenv("token", jsonStr)
		if err != nil {
			return
		}
	}
}
