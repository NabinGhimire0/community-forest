package auth

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forest-management/config"
	"forest-management/internal/models"
	"forest-management/pkg/security"
	"forest-management/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthService struct {
	db *gorm.DB
}

type LoginResult struct {
	User         *models.User
	SessionToken string
	CSRFToken    string
	ExpiresAt    time.Time
}

type MFASetupResult struct {
	Secret      string   `json:"secret"`
	OTPAuthURI  string   `json:"otpauth_uri"`
	BackupCodes []string `json:"backup_codes,omitempty"`
}

type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string { return e.Message }

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(phone, password, otp, ipAddress, userAgent string) (*LoginResult, error) {
	phone = strings.TrimSpace(phone)
	normalizedPhone, normalizeErr := security.NormalizeNepalMobile(phone)
	if normalizeErr != nil || password == "" {
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid phone or password"}
	}
	phone = normalizedPhone

	now := time.Now().UTC()
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("authentication service is unavailable")
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			tx.Rollback()
			panic(recovered)
		}
	}()

	var user models.User
	findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("phone = ?", phone).First(&user).Error
	if findErr != nil {
		_ = tx.Rollback().Error
		s.recordAudit(nil, "login_failed", "auth", nil, ipAddress, userAgent, "Invalid credentials")
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid phone or password"}
	}

	if user.Status != "active" {
		_ = tx.Rollback().Error
		s.recordAudit(&user.ID, "login_blocked", "auth", nil, ipAddress, userAgent, "Inactive account")
		return nil, &AuthError{Code: "account_inactive", Message: "This account is inactive. Contact an administrator."}
	}
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		_ = tx.Rollback().Error
		s.recordAudit(&user.ID, "login_blocked", "auth", nil, ipAddress, userAgent, "Temporarily locked account")
		return nil, &AuthError{Code: "account_locked", Message: "Too many failed attempts. Try again later."}
	}

	if !utils.CheckPassword(password, user.Password) {
		attempts := user.FailedLoginAttempts
		if user.LastFailedLoginAt == nil || now.Sub(*user.LastFailedLoginAt) > time.Duration(config.AppConfig.LoginLockMinutes)*time.Minute {
			attempts = 0
		}
		attempts++
		updates := map[string]interface{}{
			"failed_login_attempts": attempts,
			"last_failed_login_at":  now,
		}
		if attempts >= config.AppConfig.LoginMaxAttempts {
			lockedUntil := now.Add(time.Duration(config.AppConfig.LoginLockMinutes) * time.Minute)
			updates["locked_until"] = lockedUntil
		}
		_ = tx.Model(&user).Updates(updates).Error
		_ = tx.Commit().Error
		s.recordAudit(&user.ID, "login_failed", "auth", nil, ipAddress, userAgent, "Invalid credentials")
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid phone or password"}
	}

	if user.MFAEnabled {
		if strings.TrimSpace(otp) == "" {
			_ = tx.Rollback().Error
			return nil, &AuthError{Code: "mfa_required", Message: "Authenticator code is required"}
		}
		valid, err := s.verifyMFA(tx, &user, otp, now)
		if err != nil || !valid {
			_ = tx.Rollback().Error
			s.recordAudit(&user.ID, "mfa_failed", "auth", nil, ipAddress, userAgent, "Invalid MFA code")
			return nil, &AuthError{Code: "invalid_mfa", Message: "Invalid authenticator or backup code"}
		}
	}

	sessionToken, err := security.RandomToken(32)
	if err != nil {
		_ = tx.Rollback().Error
		return nil, errors.New("failed to create secure session")
	}
	csrfToken, err := security.RandomToken(32)
	if err != nil {
		_ = tx.Rollback().Error
		return nil, errors.New("failed to create secure session")
	}
	expiresAt := now.Add(time.Duration(config.AppConfig.SessionHours) * time.Hour)
	ip := nullableString(ipAddress)
	ua := nullableString(userAgent)
	session := models.UserSession{
		UserID:     user.ID,
		TokenHash:  security.SHA256Hex(sessionToken),
		CSRFHash:   security.SHA256Hex(csrfToken),
		IPAddress:  ip,
		UserAgent:  ua,
		ExpiresAt:  expiresAt,
		LastSeenAt: now,
	}
	if err := tx.Create(&session).Error; err != nil {
		_ = tx.Rollback().Error
		return nil, errors.New("failed to create session")
	}
	if err := tx.Model(&user).Updates(map[string]interface{}{
		"failed_login_attempts": 0,
		"last_failed_login_at":  nil,
		"locked_until":          nil,
		"last_login_at":         now,
	}).Error; err != nil {
		_ = tx.Rollback().Error
		return nil, errors.New("failed to update account")
	}
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to complete login")
	}

	s.recordAudit(&user.ID, "login_success", "auth", &session.ID, ipAddress, userAgent, "Secure session created")
	return &LoginResult{User: &user, SessionToken: sessionToken, CSRFToken: csrfToken, ExpiresAt: expiresAt}, nil
}

func (s *AuthService) RotateCSRF(sessionID, userID uint) (string, time.Time, error) {
	csrfToken, err := security.RandomToken(32)
	if err != nil {
		return "", time.Time{}, errors.New("failed to refresh CSRF protection")
	}
	var session models.UserSession
	if err := s.db.Where("id = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", sessionID, userID, time.Now().UTC()).First(&session).Error; err != nil {
		return "", time.Time{}, errors.New("active session not found")
	}
	if err := s.db.Model(&session).Update("csrf_hash", security.SHA256Hex(csrfToken)).Error; err != nil {
		return "", time.Time{}, errors.New("failed to refresh CSRF protection")
	}
	return csrfToken, session.ExpiresAt, nil
}

func (s *AuthService) GetProfile(userID uint) (interface{}, error) {
	var user models.User
	if err := s.db.Preload("Member.FamilyMembers").First(&user, userID).Error; err != nil {
		return nil, err
	}
	return gin.H{
		"id":                   user.ID,
		"name":                 user.Name,
		"phone":                user.Phone,
		"email":                user.Email,
		"role":                 user.Role,
		"status":               user.Status,
		"is_bootstrap_admin":   user.IsBootstrapAdmin,
		"must_change_password": user.MustChangePassword,
		"mfa_enabled":          user.MFAEnabled,
		"mfa_setup_required":   config.AppConfig.RequirePrivilegedMFA && (user.Role == "admin" || user.Role == "staff") && !user.MFAEnabled,
		"last_login_at":        user.LastLoginAt,
		"member":               user.Member,
	}, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if err := security.ValidatePassword(newPassword); err != nil {
		return err
	}
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("incorrect current password")
	}
	if utils.CheckPassword(newPassword, user.Password) {
		return errors.New("new password must be different from the current password")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"password":              hashedPassword,
			"password_changed_at":   now,
			"must_change_password":  false,
			"failed_login_attempts": 0,
			"locked_until":          nil,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
	})
}

func (s *AuthService) AdminResetPassword(userID uint, newPassword string) error {
	if err := security.ValidatePassword(newPassword); err != nil {
		return err
	}
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"password":              hashedPassword,
			"must_change_password":  true,
			"password_changed_at":   now,
			"failed_login_attempts": 0,
			"locked_until":          nil,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", now).Error
	})
}

func (s *AuthService) RevokeSession(sessionID uint, userID uint) error {
	now := time.Now().UTC()
	return s.db.Model(&models.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).
		Update("revoked_at", now).Error
}

func (s *AuthService) RevokeAllSessions(userID uint) error {
	now := time.Now().UTC()
	return s.db.Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (s *AuthService) ListSessions(userID uint) ([]models.UserSession, error) {
	var sessions []models.UserSession
	err := s.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now().UTC()).
		Order("last_seen_at DESC").Find(&sessions).Error
	return sessions, err
}

func (s *AuthService) BeginMFASetup(userID uint, currentPassword string) (*MFASetupResult, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	if !utils.CheckPassword(currentPassword, user.Password) {
		return nil, errors.New("incorrect current password")
	}
	if user.MFAEnabled {
		return nil, errors.New("MFA is already enabled; disable it before enrolling a new authenticator")
	}
	secret, err := security.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}
	encrypted, err := security.EncryptString(secret)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"mfa_secret_encrypted": encrypted,
		"mfa_enabled":          false,
		"mfa_backup_codes":     nil,
		"mfa_last_used_step":   0,
	}).Error; err != nil {
		return nil, err
	}
	issuer := "Ban Samiti"
	var settings models.SamitiSetting
	if s.db.First(&settings).Error == nil && strings.TrimSpace(settings.Name) != "" {
		issuer = settings.Name
	}
	return &MFASetupResult{Secret: secret, OTPAuthURI: security.TOTPURI(secret, user.Phone, issuer)}, nil
}

func (s *AuthService) EnableMFA(userID uint, code string) (*MFASetupResult, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	if user.MFASecretEncrypted == nil || *user.MFASecretEncrypted == "" {
		return nil, errors.New("start MFA setup first")
	}
	secret, err := security.DecryptString(*user.MFASecretEncrypted)
	if err != nil {
		return nil, err
	}
	matchedStep, valid := security.MatchTOTP(secret, code, time.Now().UTC())
	if !valid {
		return nil, errors.New("invalid authenticator code")
	}
	codes, err := security.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}
	key, err := config.EncryptionKey()
	if err != nil {
		return nil, err
	}
	hashes := make([]string, 0, len(codes))
	for _, backupCode := range codes {
		hashes = append(hashes, security.HashBackupCode(backupCode, key))
	}
	encoded, _ := json.Marshal(hashes)
	if err := s.db.Model(&user).Updates(map[string]interface{}{
		"mfa_enabled":        true,
		"mfa_backup_codes":   string(encoded),
		"mfa_last_used_step": matchedStep,
	}).Error; err != nil {
		return nil, err
	}
	return &MFASetupResult{BackupCodes: codes}, nil
}

func (s *AuthService) DisableMFA(userID uint, currentPassword, code string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPassword(currentPassword, user.Password) {
		return errors.New("incorrect current password")
	}
	valid, err := s.verifyMFA(s.db, &user, code, time.Now().UTC())
	if err != nil || !valid {
		return errors.New("invalid authenticator or backup code")
	}
	return s.db.Model(&user).Updates(map[string]interface{}{
		"mfa_enabled":          false,
		"mfa_secret_encrypted": nil,
		"mfa_backup_codes":     nil,
		"mfa_last_used_step":   0,
	}).Error
}

func (s *AuthService) verifyMFA(db *gorm.DB, user *models.User, code string, now time.Time) (bool, error) {
	if user.MFASecretEncrypted == nil {
		return false, nil
	}
	secret, err := security.DecryptString(*user.MFASecretEncrypted)
	if err != nil {
		return false, err
	}
	if matchedStep, valid := security.MatchTOTP(secret, code, now); valid {
		if matchedStep <= user.MFALastUsedStep {
			return false, nil
		}
		if err := db.Model(user).Update("mfa_last_used_step", matchedStep).Error; err != nil {
			return false, err
		}
		user.MFALastUsedStep = matchedStep
		return true, nil
	}
	if user.MFABackupCodes == nil || *user.MFABackupCodes == "" {
		return false, nil
	}
	var hashes []string
	if err := json.Unmarshal([]byte(*user.MFABackupCodes), &hashes); err != nil {
		return false, err
	}
	key, err := config.EncryptionKey()
	if err != nil {
		return false, err
	}
	candidate := security.HashBackupCode(code, key)
	for index, hash := range hashes {
		if security.ConstantTimeStringEqual(candidate, hash) {
			hashes = append(hashes[:index], hashes[index+1:]...)
			encoded, _ := json.Marshal(hashes)
			if err := db.Model(user).Update("mfa_backup_codes", string(encoded)).Error; err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *AuthService) VerifyPrivilegedStepUp(userID uint, password, code string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("authentication service is unavailable")
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			tx.Rollback()
			panic(recovered)
		}
	}()
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
		tx.Rollback()
		return errors.New("administrator account not found")
	}
	if user.Role != "admin" || user.Status != "active" || !utils.CheckPassword(password, user.Password) {
		tx.Rollback()
		return errors.New("current administrator password is incorrect")
	}
	if config.AppConfig.RequirePrivilegedMFA || user.MFAEnabled {
		if !user.MFAEnabled {
			tx.Rollback()
			return errors.New("multi-factor authentication must be enabled for this operation")
		}
		if strings.TrimSpace(code) == "" {
			tx.Rollback()
			return errors.New("a fresh authenticator or backup code is required")
		}
		valid, err := s.verifyMFA(tx, &user, code, time.Now().UTC())
		if err != nil || !valid {
			tx.Rollback()
			return errors.New("invalid or already-used authenticator code")
		}
	}
	if err := tx.Commit().Error; err != nil {
		return errors.New("could not complete security verification")
	}
	return nil
}

func (s *AuthService) VerifyCurrentPassword(userID uint, password string) error {
	var user models.User
	if err := s.db.Select("id", "password", "status").First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if user.Status != "active" || !utils.CheckPassword(password, user.Password) {
		return errors.New("current password is incorrect")
	}
	return nil
}

func (s *AuthService) recordAudit(userID *uint, action, entity string, entityID *uint, ipAddress, userAgent, remarks string) {
	ip := nullableString(ipAddress)
	ua := nullableString(userAgent)
	note := nullableString(remarks)
	_ = s.db.Create(&models.AuditLog{
		UserID: userID, Action: action, Entity: entity, EntityID: entityID,
		IPAddress: ip, UserAgent: ua, Remarks: note,
	}).Error
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func AuthErrorDetails(err error) (string, string) {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code, authErr.Message
	}
	return "authentication_failed", "Authentication failed"
}
