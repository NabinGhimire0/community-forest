package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"forest-management/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(self)")
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self' https://*.esewa.com.np")
		c.Header("Cross-Origin-Resource-Policy", "same-site")
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
			c.Header("Pragma", "no-cache")
		}
		if config.AppConfig.AppEnv == "production" {
			c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}
		c.Next()
	}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if requestID == "" || len(requestID) > 100 {
			requestID = uuid.NewString()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func LimitJSONBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

func RequireJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.Request.Method != http.MethodOptions &&
			!strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"success": false, "message": "Content-Type must be application/json"})
			return
		}
		c.Next()
	}
}

func SecureRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		_ = c.Error(fmt.Errorf("panic recovered for request %v: %v", requestID, recovered))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"message":    "An unexpected server error occurred",
			"error":      "internal_server_error",
			"request_id": requestID,
		})
	})
}
