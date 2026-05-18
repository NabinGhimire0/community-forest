package notifications

import (
	"fmt"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification creates a single notification
func (s *NotificationService) CreateNotification(userID *uint, targetRole *string, title, message, notifType string, entity *string, entityID *uint) (*models.Notification, error) {
	notif := models.Notification{
		UserID:     userID,
		TargetRole: targetRole,
		Title:      title,
		Message:    message,
		Type:       notifType,
		Entity:     entity,
		EntityID:   entityID,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(&notif).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &notif, nil
}

// NotifyUser sends a notification to a specific user
func (s *NotificationService) NotifyUser(userID uint, title, message, notifType string, entity *string, entityID *uint) (*models.Notification, error) {
	return s.CreateNotification(&userID, nil, title, message, notifType, entity, entityID)
}

// NotifyRole sends a notification to all users with a specific role
func (s *NotificationService) NotifyRole(role, title, message, notifType string, entity *string, entityID *uint) error {
	return s.db.Create(&models.Notification{
		TargetRole: &role,
		Title:      title,
		Message:    message,
		Type:       notifType,
		Entity:     entity,
		EntityID:   entityID,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}).Error
}

// Broadcast sends a notification to all users
func (s *NotificationService) Broadcast(title, message, notifType string) error {
	return s.db.Create(&models.Notification{
		Title:     title,
		Message:   message,
		Type:      notifType,
		IsRead:    false,
		CreatedAt: time.Now(),
	}).Error
}

// GetUserNotifications returns notifications for a specific user (personal + role-based + broadcast)
func (s *NotificationService) GetUserNotifications(userID uint, role string, page, perPage int) ([]models.Notification, *PaginationMeta, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).
		Where("user_id = ? OR target_role = ? OR (user_id IS NULL AND target_role IS NULL)", userID, role)

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&notifications).Error

	return notifications, &PaginationMeta{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(userID uint, role string) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("(user_id = ? OR target_role = ? OR (user_id IS NULL AND target_role IS NULL)) AND is_read = ?", userID, role, false).
		Count(&count).Error
	return count, err
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(notificationID, userID uint) error {
	now := time.Now()
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND (user_id = ? OR user_id IS NULL)", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})
	return result.Error
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uint, role string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("(user_id = ? OR target_role = ? OR (user_id IS NULL AND target_role IS NULL)) AND is_read = ?", userID, role, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// DeleteOldNotifications removes notifications older than N days
func (s *NotificationService) DeleteOldNotifications(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ? AND is_read = ?", cutoff, true).
		Delete(&models.Notification{}).Error
}
