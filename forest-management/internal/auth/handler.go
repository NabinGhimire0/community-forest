package auth

import (
	"net/http"
	"strings"
	"time"

	"forest-management/config"
	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
	OTP      string `json:"otp"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Phone and password are required")
		return
	}

	result, err := h.service.Login(req.Phone, req.Password, req.OTP, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		code, message := AuthErrorDetails(err)
		status := http.StatusUnauthorized
		if code == "mfa_required" {
			status = http.StatusPreconditionRequired
		}
		c.JSON(status, gin.H{"success": false, "message": message, "error": message, "code": code})
		return
	}

	setSessionCookies(c, result.SessionToken, result.CSRFToken, result.ExpiresAt)
	c.Header("X-CSRF-Token", result.CSRFToken)
	response.Success(c, "Login successful", gin.H{
		"user": gin.H{
			"id":                   result.User.ID,
			"name":                 result.User.Name,
			"phone":                result.User.Phone,
			"role":                 result.User.Role,
			"status":               result.User.Status,
			"is_bootstrap_admin":   result.User.IsBootstrapAdmin,
			"must_change_password": result.User.MustChangePassword,
			"mfa_enabled":          result.User.MFAEnabled,
			"mfa_setup_required":   config.AppConfig.RequirePrivilegedMFA && (result.User.Role == "admin" || result.User.Role == "staff") && !result.User.MFAEnabled,
		},
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	profile, err := h.service.GetProfile(userID)
	if err != nil {
		response.NotFound(c, "Profile not found")
		return
	}
	csrfToken, expiresAt, err := h.service.RotateCSRF(middleware.GetSessionID(c), userID)
	if err != nil {
		response.Unauthorized(c, "Your session is invalid or expired")
		return
	}
	setCSRFCookie(c, csrfToken, expiresAt)
	c.Header("X-CSRF-Token", csrfToken)
	response.Success(c, "Profile retrieved", profile)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	_ = h.service.RevokeSession(middleware.GetSessionID(c), middleware.GetUserID(c))
	clearSessionCookies(c)
	response.Success(c, "Logged out", nil)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Current and new passwords are required")
		return
	}
	if err := h.service.ChangePassword(middleware.GetUserID(c), req.OldPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "change_password", "user", &actorID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "User changed password; all sessions revoked")
	clearSessionCookies(c)
	response.Success(c, "Password changed. Sign in again with your new password.", gin.H{"reauthenticate": true})
}

type AdminResetPasswordRequest struct {
	UserID          uint   `json:"user_id" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	CurrentPassword string `json:"current_password" binding:"required"`
	MFACode         string `json:"mfa_code"`
}

func (h *AuthHandler) AdminResetPassword(c *gin.Context) {
	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "User and a strong temporary password are required")
		return
	}
	if err := h.service.VerifyPrivilegedStepUp(middleware.GetUserID(c), req.CurrentPassword, req.MFACode); err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	if err := h.service.AdminResetPassword(req.UserID, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	targetID := req.UserID
	audit.CreateAuditEntry(h.service.db, &actorID, "admin_reset_password", "user", &targetID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "Administrator reset a user password; target sessions revoked")
	response.Success(c, "Temporary password set. The user must change it at next login.", nil)
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	sessions, err := h.service.ListSessions(middleware.GetUserID(c))
	if err != nil {
		response.InternalError(c, "Could not load active sessions")
		return
	}
	response.Success(c, "Active sessions", sessions)
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	if err := h.service.RevokeAllSessions(middleware.GetUserID(c)); err != nil {
		response.InternalError(c, "Could not revoke sessions")
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "revoke_sessions", "user", &actorID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "User revoked all active sessions")
	clearSessionCookies(c)
	response.Success(c, "All sessions revoked", gin.H{"reauthenticate": true})
}

type MFABeginRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
}

func (h *AuthHandler) BeginMFA(c *gin.Context) {
	var req MFABeginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Current password is required")
		return
	}
	setup, err := h.service.BeginMFASetup(middleware.GetUserID(c), req.CurrentPassword)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, "Add this account to your authenticator app, then verify the six-digit code.", setup)
}

type MFACodeRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *AuthHandler) EnableMFA(c *gin.Context) {
	var req MFACodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Authenticator code is required")
		return
	}
	result, err := h.service.EnableMFA(middleware.GetUserID(c), req.Code)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "enable_mfa", "user", &actorID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "Multi-factor authentication enabled")
	response.Success(c, "Multi-factor authentication enabled. Store the backup codes securely; they are shown only once.", result)
}

type MFADisableRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	Code            string `json:"code" binding:"required"`
}

func (h *AuthHandler) DisableMFA(c *gin.Context) {
	var req MFADisableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Current password and an authenticator or backup code are required")
		return
	}
	if err := h.service.DisableMFA(middleware.GetUserID(c), req.CurrentPassword, req.Code); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "disable_mfa", "user", &actorID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "Multi-factor authentication disabled")
	response.Success(c, "Multi-factor authentication disabled", nil)
}

func setSessionCookies(c *gin.Context, sessionToken, csrfToken string, expiresAt time.Time) {
	maxAge := cookieMaxAge(expiresAt)
	c.SetSameSite(parseSameSite(config.AppConfig.CookieSameSite))
	c.SetCookie(config.AppConfig.SessionCookieName, sessionToken, maxAge, "/", config.AppConfig.CookieDomain, config.AppConfig.CookieSecure, true)
	setCSRFCookie(c, csrfToken, expiresAt)
}

func setCSRFCookie(c *gin.Context, csrfToken string, expiresAt time.Time) {
	c.SetSameSite(parseSameSite(config.AppConfig.CookieSameSite))
	c.SetCookie("bansamiti_csrf", csrfToken, cookieMaxAge(expiresAt), "/", config.AppConfig.CookieDomain, config.AppConfig.CookieSecure, false)
}

func cookieMaxAge(expiresAt time.Time) int {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		return 0
	}
	return maxAge
}

func clearSessionCookies(c *gin.Context) {
	c.SetSameSite(parseSameSite(config.AppConfig.CookieSameSite))
	c.SetCookie(config.AppConfig.SessionCookieName, "", -1, "/", config.AppConfig.CookieDomain, config.AppConfig.CookieSecure, true)
	c.SetCookie("bansamiti_csrf", "", -1, "/", config.AppConfig.CookieDomain, config.AppConfig.CookieSecure, false)
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
