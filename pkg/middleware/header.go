package middleware

import "github.com/gin-gonic/gin"

func GetClientIp(c *gin.Context) string {
	
	xRealIp := c.GetHeader("X-Real-IP")
	if xRealIp != "" {
		return xRealIp
	}
	
	remoteIp := c.RemoteIP()
	//remoteIp := c.GetHeader("remote_addr")
	if remoteIp != "" {
		return remoteIp
	}
	
	xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		return xForwardedFor
	}
	
	return "0.0.0.0"
}
