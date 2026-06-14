package middleware

import (
	"net/http"
	"strings"
	"time"

	"forest-management/config"
	"forest-management/database"
	"forest-management/internal/models"
	"forest-management/pkg/response"
	"forest-management/pkg/security"

	"github.com/gin-gonic/gin"
)

const (
	contextUserID    = "user_id"
	contextUserRole  = "role"
	contextPhone     = "phone"
	contextSessionID = "session_id"
	contextCSRFHash  = "csrf_hash"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := c.Cookie(config.AppConfig.SessionCookieName)
		if err != nil || strings.TrimSpace(rawToken) == "" {
			response.Unauthorized(c, "Authentication is required")
			c.Abort()
			return
		}

		now := time.Now().UTC()
		var session models.UserSession
		if database.DB == nil || database.DB.Preload("User").
			Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", security.SHA256Hex(rawToken), now).
			First(&session).Error != nil {
			response.Unauthorized(c, "Your session is invalid or expired")
			c.Abort()
			return
		}
		if session.User == nil || session.User.Status != "active" {
			response.Unauthorized(c, "Your account is inactive")
			c.Abort()
			return
		}
		if now.Sub(session.LastSeenAt) > time.Duration(config.AppConfig.SessionIdleMinutes)*time.Minute {
			_ = database.DB.Model(&session).Update("revoked_at", now).Error
			response.Unauthorized(c, "Your session expired due to inactivity")
			c.Abort()
			return
		}

		c.Set(contextUserID, session.User.ID)
		c.Set(contextPhone, session.User.Phone)
		c.Set(contextUserRole, session.User.Role)
		c.Set(contextSessionID, session.ID)
		c.Set(contextCSRFHash, session.CSRFHash)

		if now.Sub(session.LastSeenAt) >= 5*time.Minute {
			_ = database.DB.Model(&session).Update("last_seen_at", now).Error
		}

		if !safeMethod(c.Request.Method) {
			if !validCSRF(c, session.CSRFHash) {
				response.Forbidden(c, "Invalid or missing CSRF token")
				c.Abort()
				return
			}
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		if session.User.MustChangePassword && !isPasswordChangeAllowed(path) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "You must change the temporary password before continuing",
				"error":   "password_change_required",
				"code":    "password_change_required",
			})
			c.Abort()
			return
		}
		if config.AppConfig.RequirePrivilegedMFA && (session.User.Role == "admin" || session.User.Role == "staff") && !session.User.MFAEnabled && !isMFASetupAllowed(path) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Multi-factor authentication setup is required for privileged accounts",
				"error":   "mfa_setup_required",
				"code":    "mfa_setup_required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if safeMethod(c.Request.Method) {
			c.Next()
			return
		}
		hash, exists := c.Get(contextCSRFHash)
		if !exists || !validCSRF(c, hash.(string)) {
			response.Forbidden(c, "Invalid or missing CSRF token")
			c.Abort()
			return
		}
		c.Next()
	}
}

func validCSRF(c *gin.Context, expectedHash string) bool {
	cookieToken, err := c.Cookie("bansamiti_csrf")
	if err != nil || cookieToken == "" {
		return false
	}
	headerToken := c.GetHeader("X-CSRF-Token")
	if headerToken == "" || !security.ConstantTimeStringEqual(cookieToken, headerToken) {
		return false
	}
	return security.ConstantTimeStringEqual(security.SHA256Hex(headerToken), expectedHash)
}

func safeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func isPasswordChangeAllowed(path string) bool {
	return path == "/api/auth/profile" || path == "/api/auth/change-password" || path == "/api/auth/logout"
}

func isMFASetupAllowed(path string) bool {
	return path == "/api/auth/profile" || path == "/api/auth/change-password" || path == "/api/auth/logout" ||
		path == "/api/auth/mfa/setup" || path == "/api/auth/mfa/enable" || path == "/api/auth/mfa/disable"
}

func GetUserID(c *gin.Context) uint {
	value, exists := c.Get(contextUserID)
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}

func GetSessionID(c *gin.Context) uint {
	value, exists := c.Get(contextSessionID)
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}

func GetUserRole(c *gin.Context) string {
	value, exists := c.Get(contextUserRole)
	if !exists {
		return ""
	}
	role, _ := value.(string)
	return role
}
