package notifications

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func validRole(role string) bool {
	switch role {
	case "admin", "staff", "member":
		return true
	default:
		return false
	}
}

// CreateNotification creates a personal, role-targeted, or broadcast
// notification. Per-user read state is kept in notification_receipts.
func (s *NotificationService) CreateNotification(userID *uint, targetRole *string, title, message, notifType string, entity *string, entityID *uint) (*models.Notification, error) {
	title = strings.TrimSpace(title)
	message = strings.TrimSpace(message)
	notifType = strings.TrimSpace(notifType)
	if title == "" || message == "" || notifType == "" {
		return nil, errors.New("title, message and type are required")
	}
	if userID != nil && targetRole != nil {
		return nil, errors.New("notification cannot target both a user and a role")
	}
	if targetRole != nil && !validRole(*targetRole) {
		return nil, errors.New("invalid target role")
	}
	if userID != nil {
		var count int64
		if err := s.db.Model(&models.User{}).Where("id = ? AND status = ?", *userID, "active").Count(&count).Error; err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, errors.New("target user not found or inactive")
		}
	}

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

func (s *NotificationService) NotifyUser(userID uint, title, message, notifType string, entity *string, entityID *uint) (*models.Notification, error) {
	return s.CreateNotification(&userID, nil, title, message, notifType, entity, entityID)
}

func (s *NotificationService) NotifyRole(role, title, message, notifType string, entity *string, entityID *uint) error {
	_, err := s.CreateNotification(nil, &role, title, message, notifType, entity, entityID)
	return err
}

func (s *NotificationService) Broadcast(title, message, notifType string) error {
	_, err := s.CreateNotification(nil, nil, title, message, notifType, nil, nil)
	return err
}

func eligibleNotifications(db *gorm.DB, userID uint, role string) *gorm.DB {
	return db.Model(&models.Notification{}).
		Where("user_id = ? OR target_role = ? OR (user_id IS NULL AND target_role IS NULL)", userID, role)
}

func normalizePage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

// GetUserNotifications returns personal, role, and broadcast notifications,
// overlaying the authenticated user's own receipt state.
func (s *NotificationService) GetUserNotifications(userID uint, role string, page, perPage int) ([]models.Notification, *PaginationMeta, error) {
	page, perPage = normalizePage(page, perPage)
	var notifications []models.Notification
	var total int64
	query := eligibleNotifications(s.db, userID, role)
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}
	offset := (page - 1) * perPage
	if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&notifications).Error; err != nil {
		return nil, nil, err
	}

	if len(notifications) > 0 {
		ids := make([]uint, 0, len(notifications))
		for _, notification := range notifications {
			ids = append(ids, notification.ID)
		}
		var receipts []models.NotificationReceipt
		if err := s.db.Where("user_id = ? AND notification_id IN ?", userID, ids).Find(&receipts).Error; err != nil {
			return nil, nil, err
		}
		receiptByNotification := make(map[uint]models.NotificationReceipt, len(receipts))
		for _, receipt := range receipts {
			receiptByNotification[receipt.NotificationID] = receipt
		}
		for index := range notifications {
			notifications[index].IsRead = false
			notifications[index].ReadAt = nil
			if receipt, ok := receiptByNotification[notifications[index].ID]; ok {
				notifications[index].IsRead = receipt.IsRead
				notifications[index].ReadAt = receipt.ReadAt
			}
		}
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	return notifications, &PaginationMeta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, nil
}

func (s *NotificationService) GetUnreadCount(userID uint, role string) (int64, error) {
	var count int64
	err := eligibleNotifications(s.db, userID, role).
		Where("NOT EXISTS (?)", s.db.Model(&models.NotificationReceipt{}).
			Select("1").
			Where("notification_receipts.notification_id = notifications.id AND notification_receipts.user_id = ? AND notification_receipts.is_read = ?", userID, true)).
		Count(&count).Error
	return count, err
}

func (s *NotificationService) notificationIsEligible(notificationID, userID uint, role string) (bool, error) {
	var count int64
	err := eligibleNotifications(s.db, userID, role).Where("notifications.id = ?", notificationID).Count(&count).Error
	return count > 0, err
}

func (s *NotificationService) MarkAsRead(notificationID, userID uint, role string) error {
	eligible, err := s.notificationIsEligible(notificationID, userID, role)
	if err != nil {
		return err
	}
	if !eligible {
		return gorm.ErrRecordNotFound
	}
	now := time.Now()
	receipt := models.NotificationReceipt{
		NotificationID: notificationID,
		UserID:         userID,
		IsRead:         true,
		ReadAt:         &now,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "notification_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}),
	}).Create(&receipt).Error
}

func (s *NotificationService) MarkAllAsRead(userID uint, role string) error {
	var ids []uint
	if err := eligibleNotifications(s.db, userID, role).Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	receipts := make([]models.NotificationReceipt, 0, len(ids))
	for _, id := range ids {
		receipts = append(receipts, models.NotificationReceipt{
			NotificationID: id,
			UserID:         userID,
			IsRead:         true,
			ReadAt:         &now,
		})
	}
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "notification_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"is_read":    true,
			"read_at":    now,
			"updated_at": now,
		}),
	}).CreateInBatches(receipts, 500).Error
}

// DeleteOldNotifications removes old notifications only after every currently
// eligible active user has a read receipt. It is deliberately conservative.
func (s *NotificationService) DeleteOldNotifications(days int) error {
	if days < 30 {
		days = 30
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Exec(`
		DELETE FROM notifications n
		WHERE n.created_at < ?
		AND NOT EXISTS (
			SELECT 1 FROM users u
			WHERE u.status = 'active'
			AND (n.user_id = u.id OR n.target_role = u.role OR (n.user_id IS NULL AND n.target_role IS NULL))
			AND NOT EXISTS (
				SELECT 1 FROM notification_receipts nr
				WHERE nr.notification_id = n.id AND nr.user_id = u.id AND nr.is_read = true
			)
		)
	`, cutoff).Error
}
