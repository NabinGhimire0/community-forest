package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
)

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// LogEntry creates an audit log record
func (s *AuditService) LogEntry(userID *uint, action, entity string, entityID *uint, oldValues, newValues interface{}, ipAddress, userAgent, remarks string) error {
	var oldJSON, newJSON *string

	oldJSON = safeAuditJSON(oldValues)
	newJSON = safeAuditJSON(newValues)

	log := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		OldValues: oldJSON,
		NewValues: newJSON,
		IPAddress: &ipAddress,
		UserAgent: &userAgent,
		Remarks:   &remarks,
		CreatedAt: time.Now(),
	}

	return s.db.Create(&log).Error
}

func safeAuditJSON(value interface{}) *string {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var normalized interface{}
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil
	}
	redactAuditSecrets(normalized)
	encoded, err = json.Marshal(normalized)
	if err != nil {
		return nil
	}
	text := string(encoded)
	return &text
}

func redactAuditSecrets(value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "password") || strings.Contains(lower, "passphrase") ||
				strings.Contains(lower, "secret") || strings.Contains(lower, "token") ||
				strings.Contains(lower, "mfa_code") || strings.Contains(lower, "otp") ||
				strings.Contains(lower, "backup_code") || strings.Contains(lower, "gateway_response") {
				typed[key] = "[REDACTED]"
				continue
			}
			redactAuditSecrets(child)
		}
	case []interface{}:
		for _, child := range typed {
			redactAuditSecrets(child)
		}
	}
}

// ListLogs returns paginated audit logs with filters
func (s *AuditService) ListLogs(page, perPage int, action, entity, userID string) ([]models.AuditLog, *PaginationMeta, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.Model(&models.AuditLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if entity != "" {
		query = query.Where("entity = ?", entity)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&logs).Error

	return logs, &PaginationMeta{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

// GetEntityHistory returns all audit logs for a specific entity
func (s *AuditService) GetEntityHistory(entity string, entityID uint) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.Preload("User").
		Where("entity = ? AND entity_id = ?", entity, entityID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Helper to create audit entries from gin context
func CreateAuditEntry(db *gorm.DB, userID *uint, action, entity string, entityID *uint, oldValues, newValues interface{}, ipAddress, userAgent, remarks string) {
	service := NewAuditService(db)
	if err := service.LogEntry(userID, action, entity, entityID, oldValues, newValues, ipAddress, userAgent, remarks); err != nil {
		fmt.Printf("⚠️ Failed to create audit log: %v\n", err)
	}
}
