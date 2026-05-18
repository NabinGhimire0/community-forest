package auth

import (
	"errors"
	"fmt"

	"forest-management/internal/models"
	"forest-management/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

// Login authenticates a user by phone and password
func (s *AuthService) Login(phone, password string) (string, *models.User, error) {
	// 1. Find user by phone
	var user models.User
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return "", nil, errors.New("invalid phone or password")
	}

	// 2. Check if user is active
	if user.Status != "active" {
		return "", nil, errors.New("your account is inactive. Contact admin.")
	}

	// 3. Compare password
	if !utils.CheckPassword(password, user.Password) {
		return "", nil, errors.New("invalid phone or password")
	}

	// 4. Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Phone, user.Role)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, &user, nil
}

// GetProfile returns the member profile associated with a user
func (s *AuthService) GetProfile(userID uint) (interface{}, error) {
	var user models.User
	if err := s.db.Preload("Member").First(&user, userID).Error; err != nil {
		return nil, err
	}

	return gin.H{
		"id":     user.ID,
		"name":   user.Name,
		"phone":  user.Phone,
		"email":  user.Email,
		"role":   user.Role,
		"member": user.Member,
	}, nil
}

// ChangePassword changes the password for a user (requires old password)
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("incorrect old password")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// AdminResetPassword resets a user's password (admin only, no old password required)
func (s *AuthService) AdminResetPassword(userID uint, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.db.Model(&user).Update("password", hashedPassword).Error
}
