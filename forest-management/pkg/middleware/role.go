package middleware

import (
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequireRole checks if the authenticated user has the required role
// Usage: router.Use(middleware.RequireRole("admin"))
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		// Check if user's role is in the allowed roles list
		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "You do not have permission to access this resource")
			c.Abort()
			return
		}

		c.Next()
	}
}
