package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort    string
	AppEnv     string
	AppVersion string

	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DBMaxOpenConns   int
	DBMaxIdleConns   int
	DBConnMaxMinutes int
	RunAutoMigrate   bool
	AllowInsecureDB  bool

	CORSOrigins      string
	FrontendURL      string
	PublicBackendURL string
	TrustedProxies   string

	SessionCookieName    string
	CookieDomain         string
	CookieSecure         bool
	CookieSameSite       string
	SessionHours         int
	SessionIdleMinutes   int
	DataEncryptionKey    string
	RequirePrivilegedMFA bool

	LoginMaxAttempts   int
	LoginLockMinutes   int
	RateLimitPerMinute int
	MaxJSONBodyBytes   int64

	UploadDir        string
	ClamAVAddr       string
	RequireAntivirus bool

	PGDumpPath           string
	BackupTempDir        string
	BackupTimeoutMinutes int
	ExportMaxRows        int

	EsewaMode        string
	EsewaProductCode string
	EsewaSecretKey   string
	EsewaFormURL     string
	EsewaStatusURL   string

	SMSProvider string
	SMSAPIKey   string
	SMSSenderID string

	SeedAdminName     string
	SeedAdminPhone    string
	SeedAdminPassword string

	SeedSamitiName           string
	SeedSamitiRegistrationNo string
	SeedSamitiAddress        string
	SeedSamitiWardNo         string
	SeedSamitiMunicipality   string
	SeedSamitiDistrict       string
	SeedSamitiProvince       string
}

var AppConfig *Config

func InitConfig() {
	_ = godotenv.Load()

	appEnv := strings.ToLower(getEnv("APP_ENV", "development"))
	runAutoMigrateDefault := "true"
	cookieSecureDefault := "false"
	sessionCookieNameDefault := "bansamiti_session"
	if appEnv == "production" {
		runAutoMigrateDefault = "false"
		cookieSecureDefault = "true"
		sessionCookieNameDefault = "__Host-bansamiti_session"
	}

	AppConfig = &Config{
		AppPort:                  getEnv("APP_PORT", "8080"),
		AppEnv:                   appEnv,
		AppVersion:               getEnv("APP_VERSION", "1.0.0"),
		DBHost:                   getEnv("DB_HOST", "localhost"),
		DBPort:                   getEnv("DB_PORT", "5432"),
		DBUser:                   getEnv("DB_USER", "postgres"),
		DBPassword:               getEnv("DB_PASSWORD", "admin123"),
		DBName:                   getEnv("DB_NAME", "forest_management"),
		DBSSLMode:                getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:           getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxMinutes:         getEnvInt("DB_CONN_MAX_MINUTES", 30),
		RunAutoMigrate:           getEnvBool("RUN_AUTO_MIGRATE", runAutoMigrateDefault == "true"),
		AllowInsecureDB:          getEnvBool("ALLOW_INSECURE_DB", false),
		CORSOrigins:              getEnv("CORS_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"),
		FrontendURL:              strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:5173"), "/"),
		PublicBackendURL:         strings.TrimRight(getEnv("PUBLIC_BACKEND_URL", "http://localhost:8080"), "/"),
		TrustedProxies:           getEnv("TRUSTED_PROXIES", ""),
		SessionCookieName:        getEnv("SESSION_COOKIE_NAME", sessionCookieNameDefault),
		CookieDomain:             getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:             getEnvBool("COOKIE_SECURE", cookieSecureDefault == "true"),
		CookieSameSite:           strings.ToLower(getEnv("COOKIE_SAME_SITE", "lax")),
		SessionHours:             getEnvInt("SESSION_HOURS", 12),
		SessionIdleMinutes:       getEnvInt("SESSION_IDLE_MINUTES", 60),
		DataEncryptionKey:        getEnv("DATA_ENCRYPTION_KEY", ""),
		RequirePrivilegedMFA:     getEnvBool("REQUIRE_PRIVILEGED_MFA", appEnv == "production"),
		LoginMaxAttempts:         getEnvInt("LOGIN_MAX_ATTEMPTS", 5),
		LoginLockMinutes:         getEnvInt("LOGIN_LOCK_MINUTES", 15),
		RateLimitPerMinute:       getEnvInt("RATE_LIMIT_PER_MINUTE", 180),
		MaxJSONBodyBytes:         getEnvInt64("MAX_JSON_BODY_BYTES", 2*1024*1024),
		UploadDir:                getEnv("UPLOAD_DIR", "./uploads"),
		ClamAVAddr:               getEnv("CLAMAV_ADDR", ""),
		RequireAntivirus:         getEnvBool("REQUIRE_ANTIVIRUS", false),
		PGDumpPath:               getEnv("PG_DUMP_PATH", "pg_dump"),
		BackupTempDir:            getEnv("BACKUP_TEMP_DIR", ""),
		BackupTimeoutMinutes:     getEnvInt("BACKUP_TIMEOUT_MINUTES", 15),
		ExportMaxRows:            getEnvInt("EXPORT_MAX_ROWS", 250000),
		EsewaMode:                strings.ToLower(getEnv("ESEWA_MODE", "test")),
		EsewaProductCode:         getEnv("ESEWA_PRODUCT_CODE", "EPAYTEST"),
		EsewaSecretKey:           getEnv("ESEWA_SECRET_KEY", "8gBm/:&EnhH.1/q"),
		EsewaFormURL:             getEnv("ESEWA_FORM_URL", "https://rc-epay.esewa.com.np/api/epay/main/v2/form"),
		EsewaStatusURL:           getEnv("ESEWA_STATUS_URL", "https://rc.esewa.com.np/api/epay/transaction/status/"),
		SMSProvider:              getEnv("SMS_PROVIDER", "disabled"),
		SMSAPIKey:                getEnv("SMS_API_KEY", ""),
		SMSSenderID:              getEnv("SMS_SENDER_ID", "BanSamiti"),
		SeedAdminName:            getEnv("SEED_ADMIN_NAME", "System Admin"),
		SeedAdminPhone:           getEnv("SEED_ADMIN_PHONE", ""),
		SeedAdminPassword:        getEnv("SEED_ADMIN_PASSWORD", ""),
		SeedSamitiName:           getEnv("SEED_SAMITI_NAME", ""),
		SeedSamitiRegistrationNo: getEnv("SEED_SAMITI_REGISTRATION_NO", ""),
		SeedSamitiAddress:        getEnv("SEED_SAMITI_ADDRESS", ""),
		SeedSamitiWardNo:         getEnv("SEED_SAMITI_WARD_NO", ""),
		SeedSamitiMunicipality:   getEnv("SEED_SAMITI_MUNICIPALITY", ""),
		SeedSamitiDistrict:       getEnv("SEED_SAMITI_DISTRICT", ""),
		SeedSamitiProvince:       getEnv("SEED_SAMITI_PROVINCE", ""),
	}

	if err := validate(); err != nil {
		panic(err)
	}
	fmt.Println("✅ Configuration loaded")
}

func validate() error {
	cfg := AppConfig
	if cfg.SessionHours < 1 || cfg.SessionHours > 168 {
		return fmt.Errorf("SESSION_HOURS must be between 1 and 168")
	}
	if cfg.SessionIdleMinutes < 5 || cfg.SessionIdleMinutes > cfg.SessionHours*60 {
		return fmt.Errorf("SESSION_IDLE_MINUTES must be between 5 and SESSION_HOURS")
	}
	if cfg.LoginMaxAttempts < 3 || cfg.LoginMaxAttempts > 20 {
		return fmt.Errorf("LOGIN_MAX_ATTEMPTS must be between 3 and 20")
	}
	if cfg.CookieSameSite != "lax" && cfg.CookieSameSite != "strict" && cfg.CookieSameSite != "none" {
		return fmt.Errorf("COOKIE_SAME_SITE must be lax, strict, or none")
	}
	if cfg.CookieSameSite == "none" && !cfg.CookieSecure {
		return fmt.Errorf("COOKIE_SECURE must be true when COOKIE_SAME_SITE=none")
	}
	if cfg.DataEncryptionKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(cfg.DataEncryptionKey)
		if err != nil || len(decoded) != 32 {
			return fmt.Errorf("DATA_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
		}
	}
	if cfg.RequireAntivirus && strings.TrimSpace(cfg.ClamAVAddr) == "" {
		return fmt.Errorf("CLAMAV_ADDR is required when REQUIRE_ANTIVIRUS=true")
	}
	if cfg.EsewaMode == "live" && (cfg.EsewaProductCode == "" || cfg.EsewaSecretKey == "") {
		return fmt.Errorf("ESEWA_PRODUCT_CODE and ESEWA_SECRET_KEY are required in live mode")
	}

	if cfg.AppEnv != "production" {
		return nil
	}
	if cfg.DBPassword == "" || cfg.DBPassword == "admin123" || len(cfg.DBPassword) < 12 {
		return fmt.Errorf("production DB_PASSWORD must be a strong non-default password")
	}
	if strings.EqualFold(cfg.DBSSLMode, "disable") && !cfg.AllowInsecureDB {
		return fmt.Errorf("production DB_SSLMODE cannot be disable unless ALLOW_INSECURE_DB=true is explicitly set")
	}
	if cfg.CookieSecure == false {
		return fmt.Errorf("COOKIE_SECURE must be true in production")
	}
	if !strings.HasPrefix(cfg.SessionCookieName, "__Host-") || strings.TrimSpace(cfg.CookieDomain) != "" {
		return fmt.Errorf("production SESSION_COOKIE_NAME must use the __Host- prefix and COOKIE_DOMAIN must be empty")
	}
	if !cfg.RequirePrivilegedMFA {
		return fmt.Errorf("REQUIRE_PRIVILEGED_MFA must be true in production")
	}
	if !cfg.RequireAntivirus || strings.TrimSpace(cfg.ClamAVAddr) == "" {
		return fmt.Errorf("production file uploads require REQUIRE_ANTIVIRUS=true and CLAMAV_ADDR")
	}
	if !filepath.IsAbs(cfg.UploadDir) || !filepath.IsAbs(cfg.BackupTempDir) || !filepath.IsAbs(cfg.PGDumpPath) {
		return fmt.Errorf("production UPLOAD_DIR, BACKUP_TEMP_DIR, and PG_DUMP_PATH must be absolute paths")
	}
	if filepath.Clean(cfg.UploadDir) == filepath.Clean(cfg.BackupTempDir) {
		return fmt.Errorf("UPLOAD_DIR and BACKUP_TEMP_DIR must be different directories")
	}
	if cfg.DataEncryptionKey == "" {
		return fmt.Errorf("DATA_ENCRYPTION_KEY is required in production")
	}
	if containsWildcardOrigin(cfg.CORSOrigins) {
		return fmt.Errorf("CORS_ORIGINS cannot contain * in production")
	}
	if !isHTTPSURL(cfg.FrontendURL) || !isHTTPSURL(cfg.PublicBackendURL) {
		return fmt.Errorf("FRONTEND_URL and PUBLIC_BACKEND_URL must use HTTPS in production")
	}
	if cfg.RunAutoMigrate {
		return fmt.Errorf("RUN_AUTO_MIGRATE must be false in production; run the migration command during deployment")
	}
	if cfg.EsewaMode != "live" {
		return fmt.Errorf("ESEWA_MODE must be live in production")
	}
	if cfg.EsewaProductCode == "EPAYTEST" || strings.Contains(cfg.EsewaFormURL, "rc-epay") || strings.Contains(cfg.EsewaStatusURL, "rc.esewa") {
		return fmt.Errorf("eSewa test credentials and endpoints cannot be used in production")
	}
	return nil
}

func EncryptionKey() ([]byte, error) {
	if AppConfig == nil || AppConfig.DataEncryptionKey == "" {
		return nil, fmt.Errorf("DATA_ENCRYPTION_KEY is not configured")
	}
	key, err := base64.StdEncoding.DecodeString(AppConfig.DataEncryptionKey)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("invalid DATA_ENCRYPTION_KEY")
	}
	return key, nil
}

func containsWildcardOrigin(value string) bool {
	for _, origin := range strings.Split(value, ",") {
		if strings.TrimSpace(origin) == "*" {
			return true
		}
	}
	return false
}

func isHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && strings.EqualFold(parsed.Scheme, "https") && parsed.Host != ""
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt64(key string, fallback int64) int64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
