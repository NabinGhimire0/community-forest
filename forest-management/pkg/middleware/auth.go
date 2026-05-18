package middleware

import (
	"strings"

	"forest-management/pkg/response"
	"forest-management/pkg/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from Authorization header
// If valid, it extracts user info and puts it in gin context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort() // Stop the request from continuing
			return
		}

		// 2. Check "Bearer " prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format. Use: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Parse and validate the token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// 4. Extract user info from claims and store in context
		// Context values are available to all downstream handlers
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("phone", claims["phone"].(string))
		c.Set("role", claims["role"].(string))

		// 5. Continue to the next handler
		c.Next()
	}
}

// GetUserID is a helper to extract user_id from gin context
func GetUserID(c *gin.Context) uint {
	id, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return id.(uint)
}

// GetUserRole extracts role from gin context
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("role")
	if !exists {
		return ""
	}
	return role.(string)
}
