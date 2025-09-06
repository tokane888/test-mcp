package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 機密情報のフィールド名
var sensitiveFields = map[string]bool{
	"password":      true,
	"password_hash": true,
	"x-api-key":     true,
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// リクエストボディを読み込む
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// レスポンスをキャプチャするためのカスタムWriter
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// リクエスト処理
		c.Next()

		// ログフィールド
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", time.Since(start)),
		}

		// リクエストヘッダー（機密情報をマスク）
		maskedHeaders := maskHeaders(c.Request.Header)
		fields = append(fields, zap.Any("request_headers", maskedHeaders))

		// リクエストボディ（機密情報をマスク）
		if len(requestBody) > 0 && isJSON(c.GetHeader("Content-Type")) {
			maskedBody := maskJSON(requestBody)
			fields = append(fields, zap.Any("request_body", maskedBody))
		}

		// レスポンスボディ（エラーレスポンスのみ記録）
		if c.Writer.Status() >= 400 && blw.body.Len() > 0 {
			fields = append(fields, zap.String("response_body", blw.body.String()))
		}

		// ログレベルの決定
		switch {
		case c.Writer.Status() >= 500:
			logger.Error("request failed", fields...)
		case c.Writer.Status() >= 400:
			logger.Warn("request error", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}

func maskHeaders(headers map[string][]string) map[string][]string {
	masked := make(map[string][]string)
	for k, v := range headers {
		if sensitiveFields[strings.ToLower(k)] {
			masked[k] = []string{"***MASKED***"}
		} else {
			masked[k] = v
		}
	}
	return masked
}

func maskJSON(data []byte) any {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return string(data)
	}
	maskValue(obj)
	return obj
}

func maskValue(v any) {
	switch val := v.(type) {
	case map[string]any:
		for k, v := range val {
			if sensitiveFields[strings.ToLower(k)] {
				val[k] = "***MASKED***"
			} else {
				maskValue(v)
			}
		}
	case []any:
		for _, item := range val {
			maskValue(item)
		}
	}
}

func isJSON(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}
