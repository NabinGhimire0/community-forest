package auth

import (
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// LoginRequest is the input DTO (Data Transfer Object)
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"` // Phone is the username
	Password string `json:"password" binding:"required"`
}

// Login handles member login
// @Summary Member login
// @Description Authenticates a member using phone and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Phone and password are required")
		return
	}

	token, user, err := h.service.Login(req.Phone, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, "Login successful", gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"phone": user.Phone,
			"role":  user.Role,
		},
	})
}

// GetProfile returns the authenticated user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// user_id is injected by AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	profile, err := h.service.GetProfile(userID.(uint))
	if err != nil {
		response.NotFound(c, "Profile not found")
		return
	}

	response.Success(c, "Profile retrieved", profile)
}

// ChangePasswordRequest is the DTO for password change
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword handles password change for the logged-in user
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Old password and new password (min 6 chars) are required")
		return
	}

	err := h.service.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	response.Success(c, "Password changed successfully", nil)
}

// AdminResetPassword allows admin to reset any user's password
type AdminResetPasswordRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) AdminResetPassword(c *gin.Context) {
	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	err := h.service.AdminResetPassword(req.UserID, req.NewPassword)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Password reset successfully", nil)
}
