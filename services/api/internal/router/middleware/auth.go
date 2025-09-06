package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tokane888/go-repository-template/services/api/internal/dto/response"
)

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, response.NewError("MISSING_API_KEY", "API Keyが指定されていません"))
			c.Abort()
			return
		}

		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			// 環境変数が設定されていない場合はエラー
			c.JSON(http.StatusInternalServerError, response.NewError("CONFIG_ERROR", "API Key設定エラー"))
			c.Abort()
			return
		}

		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, response.NewError("INVALID_API_KEY", "無効なAPI Keyです"))
			c.Abort()
			return
		}

		c.Next()
	}
}
