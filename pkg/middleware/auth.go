package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/startopsz/rule/pkg/response/errCode"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"regexp"
	
	"github.com/gin-gonic/gin"
	"time"
)

/*
OpenTelemetry
*/

func OpenTelemetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 使用 OpenTelemetry 发送数据
		
		// 2. 获取 tranceId 并放在
	}
}

/*
token

验证 token 参数是否存在
*/

func Authorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		
		m, err := regexp.Match("/swagger", []byte(c.Request.RequestURI))
		if err != nil {
			c.JSON(401, gin.H{"errCode": errCode.UserUnAuthorizeCode, "errMsg": errCode.UserUnAuthorizeMsg})
			c.Abort()
			return
		}
		if m {
			return
		}
		
		token := c.GetHeader("token")
		
		if token == "" {
			token = c.Query("token")
		}
		
		if token == "" {
			c.JSON(401, gin.H{"errCode": errCode.UserUnAuthorizeCode, "errMsg": errCode.UserUnAuthorizeMsg})
			c.Abort()
			return
		}
		
		c.Set("token", token)
	}
}

func Recording(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		traceId := trace.SpanContextFromContext(c.Request.Context()).TraceID().String()
		spanId := trace.SpanContextFromContext(c.Request.Context()).SpanID().String()
		c.Header("Request-Id", traceId)
		// 设置 error 默认值为 ok
		c.Set("error", "ok")
		c.Next()
		
		err, _ := c.Get("error")
		
		logger.Info(err.(string),
			zap.String("traceId", traceId),
			zap.String("spanId", spanId),
			zap.Int64("startTimestamp", start.Unix()),
			zap.String("token", c.GetHeader("token")),
			zap.String("clientIP", c.ClientIP()),
			zap.String("requestURI", c.Request.RequestURI),
			zap.String("contentType", c.ContentType()),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("form", c.Request.Form.Encode()),
			//zap.String("traceId", c.Request.Header.),
			zap.Float64("latency", time.Now().Sub(start).Seconds()),
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

func Sign(source []byte, salt []byte) (string, error) {
	
	mac := hmac.New(sha256.New, salt)
	_, err := mac.Write(source)
	
	r := hex.EncodeToString(mac.Sum(nil))
	
	if err != nil {
		return "", err
	}
	return r, nil
}
