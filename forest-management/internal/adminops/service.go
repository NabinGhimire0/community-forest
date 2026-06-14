package adminops

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"forest-management/config"
	"forest-management/internal/auth"
	"forest-management/pkg/backupcrypto"

	"gorm.io/gorm"
)

type Service struct {
	db       *gorm.DB
	backupMu sync.Mutex
	exportMu sync.Mutex
}

type ExportFilter struct {
	FiscalYearID uint
	FromDate     string
	ToDate       string
}

type datasetDefinition struct {
	Name         string
	Filename     string
	SQL          string
	FiscalColumn string
	DateColumn   string
}

type backupManifest struct {
	FormatVersion string            `json:"format_version"`
	Application   string            `json:"application"`
	AppVersion    string            `json:"app_version"`
	CreatedAt     string            `json:"created_at"`
	Database      string            `json:"database"`
	IncludesFiles bool              `json:"includes_files"`
	Files         map[string]string `json:"sha256"`
	RestoreNotes  []string          `json:"restore_notes"`
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

var datasets = map[string]datasetDefinition{
	"members": {
		Name: "members", Filename: "members.csv", DateColumn: "m.created_at",
		SQL: `SELECT m.membership_no, m.name, m.assistant_name, m.father_name, m.ward_no, m.tole,
		COALESCE(m.phone, '') AS phone, m.joined_date, m.status, COALESCE(m.remarks, '') AS remarks,
		m.created_at, m.updated_at FROM members m WHERE m.deleted_at IS NULL`,
	},
	"family_members": {
		Name: "family_members", Filename: "family_members.csv", DateColumn: "fm.created_at",
		SQL: `SELECT m.membership_no, m.name AS household_head, fm.name, fm.relation, fm.age,
		COALESCE(fm.gender, '') AS gender, COALESCE(fm.citizenship_no, '') AS citizenship_no,
		COALESCE(fm.remarks, '') AS remarks, fm.created_at
		FROM family_members fm JOIN members m ON m.id = fm.member_id WHERE m.deleted_at IS NULL`,
	},
	"membership_fees": {
		Name: "membership_fees", Filename: "membership_fees.csv", FiscalColumn: "t.fiscal_year_id", DateColumn: "t.date",
		SQL: `SELECT m.membership_no, m.name AS member_name, fy.name AS fiscal_year, t.type, t.source,
		t.record_status, t.total_amount, t.amount_paid, t.amount_remaining, t.receipt_no,
		COALESCE(t.physical_reference, '') AS physical_reference, t.date, COALESCE(t.remarks, '') AS remarks
		FROM transactions t JOIN members m ON m.id = t.member_id JOIN fiscal_years fy ON fy.id = t.fiscal_year_id
		WHERE t.type IN ('membership_fee', 'legacy_gasti_fee')`,
	},
	"sales": {
		Name: "sales", Filename: "sales.csv", FiscalColumn: "t.fiscal_year_id", DateColumn: "t.date",
		SQL: `SELECT m.membership_no, m.name AS member_name, fy.name AS fiscal_year, t.type, t.source,
		COALESCE(ri.name, '') AS resource_item, COALESCE(t.quantity, 0) AS quantity,
		COALESCE(t.rate_per_unit, 0) AS rate_per_unit, t.total_amount, t.amount_paid, t.amount_remaining,
		t.receipt_no, t.record_status, t.date, COALESCE(t.remarks, '') AS remarks
		FROM transactions t JOIN members m ON m.id = t.member_id JOIN fiscal_years fy ON fy.id = t.fiscal_year_id
		LEFT JOIN resource_items ri ON ri.id = t.resource_item_id
		WHERE t.type IN ('resource_sale', 'legacy_timber_sale', 'legacy_firewood_sale', 'legacy_other_sale')`,
	},
	"transactions": {
		Name: "transactions", Filename: "transactions.csv", FiscalColumn: "t.fiscal_year_id", DateColumn: "t.date",
		SQL: `SELECT t.id, m.membership_no, m.name AS member_name, fy.name AS fiscal_year, t.type, t.source,
		t.record_status, t.total_amount, t.amount_paid, t.amount_remaining, t.receipt_no,
		COALESCE(t.physical_reference, '') AS physical_reference, t.date, COALESCE(t.remarks, '') AS remarks,
		t.created_at, t.updated_at
		FROM transactions t JOIN members m ON m.id = t.member_id JOIN fiscal_years fy ON fy.id = t.fiscal_year_id
		WHERE 1=1`,
	},
	"payments": {
		Name: "payments", Filename: "payments.csv", DateColumn: "p.created_at",
		SQL: `SELECT p.id, m.membership_no, m.name AS member_name, p.amount, p.payment_method, p.status,
		COALESCE(p.transaction_id, '') AS external_transaction_id, COALESCE(p.gateway_reference_id, '') AS gateway_reference_id,
		COALESCE(p.gateway_status, '') AS gateway_status, COALESCE(p.receipt_no, '') AS receipt_no,
		p.paid_at, p.verified_at, p.created_at, COALESCE(p.remarks, '') AS remarks
		FROM payments p JOIN members m ON m.id = p.member_id WHERE 1=1`,
	},
	"requests": {
		Name: "requests", Filename: "requests.csv", FiscalColumn: "r.fiscal_year_id", DateColumn: "r.requested_at",
		SQL: `SELECT r.id, m.membership_no, m.name AS member_name, fy.name AS fiscal_year, ri.name AS resource_item,
		r.quantity_requested, COALESCE(r.quantity_approved, 0) AS quantity_approved,
		COALESCE(r.rate_per_unit, 0) AS rate_per_unit, COALESCE(r.total_amount, 0) AS total_amount,
		r.status, r.requested_at, r.approved_at, COALESCE(r.remarks, '') AS remarks
		FROM requests r JOIN members m ON m.id = r.member_id JOIN fiscal_years fy ON fy.id = r.fiscal_year_id
		JOIN resource_items ri ON ri.id = r.resource_item_id WHERE 1=1`,
	},
	"stock": {
		Name: "stock", Filename: "stock.csv", FiscalColumn: "s.fiscal_year_id",
		SQL: `SELECT fy.name AS fiscal_year, rt.name AS resource_type, ri.name AS resource_item, rt.unit,
		s.total_quantity, s.remaining_quantity, s.reserved_quantity,
		(s.remaining_quantity - s.reserved_quantity) AS available_quantity
		FROM stocks s JOIN fiscal_years fy ON fy.id = s.fiscal_year_id
		JOIN resource_items ri ON ri.id = s.resource_item_id
		JOIN resource_types rt ON rt.id = ri.resource_type_id WHERE 1=1`,
	},
	"expenses": {
		Name: "expenses", Filename: "expenses.csv", FiscalColumn: "e.fiscal_year_id", DateColumn: "e.expense_date",
		SQL: `SELECT e.id, fy.name AS fiscal_year, ec.name AS category, e.title, e.amount, e.expense_date,
		e.payment_method, e.paid_to, COALESCE(e.receipt_no, '') AS receipt_no,
		COALESCE(e.remarks, '') AS remarks, u.name AS created_by, e.created_at
		FROM expenses e JOIN fiscal_years fy ON fy.id = e.fiscal_year_id
		JOIN expense_categories ec ON ec.id = e.category_id JOIN users u ON u.id = e.created_by WHERE e.deleted_at IS NULL`,
	},
	"fines": {
		Name: "fines", Filename: "fines.csv", FiscalColumn: "f.fiscal_year_id", DateColumn: "f.incident_date",
		SQL: `SELECT f.id, fy.name AS fiscal_year, COALESCE(m.membership_no, '') AS membership_no,
		COALESCE(m.name, f.name) AS person_name, f.violation_type, COALESCE(f.description, '') AS description,
		f.fine_amount, f.incident_date, f.status, COALESCE(f.payment_reference, '') AS payment_reference,
		COALESCE(f.remarks, '') AS remarks, f.created_at
		FROM fines f JOIN fiscal_years fy ON fy.id = f.fiscal_year_id LEFT JOIN members m ON m.id = f.member_id WHERE f.deleted_at IS NULL`,
	},
	"letters": {
		Name: "letters", Filename: "letters.csv", DateColumn: "l.letter_date",
		SQL: `SELECT l.id, l.type, COALESCE(l.reference_no, '') AS reference_no, l.title, l.subject,
		COALESCE(l.from_party, '') AS from_party, COALESCE(l.to_party, '') AS to_party,
		l.letter_date, l.received_date, l.sent_date, COALESCE(l.remarks, '') AS remarks, l.created_at
		FROM letters l WHERE l.status = 'active'`,
	},
	"committee": {
		Name: "committee", Filename: "committee.csv", DateColumn: "h.created_at",
		SQL: `SELECT h.name, h.post, COALESCE(h.phone, '') AS phone, COALESCE(h.email, '') AS email,
		COALESCE(h.address, '') AS address, h.tenure_start, h.tenure_end, h.is_active,
		COALESCE(u.role, '') AS system_role, COALESCE(h.remarks, '') AS remarks
		FROM samiti_heads h LEFT JOIN users u ON u.id = h.user_id WHERE 1=1`,
	},
	"audit_logs": {
		Name: "audit_logs", Filename: "audit_logs.csv", DateColumn: "a.created_at",
		SQL: `SELECT a.id, COALESCE(u.name, 'System') AS user_name, COALESCE(u.phone, '') AS user_phone,
		a.action, a.entity, a.entity_id, COALESCE(a.ip_address, '') AS ip_address,
		COALESCE(a.user_agent, '') AS user_agent, COALESCE(a.remarks, '') AS remarks, a.created_at
		FROM audit_logs a LEFT JOIN users u ON u.id = a.user_id WHERE 1=1`,
	},
}

func (s *Service) VerifyAdminCredentials(userID uint, password, mfaCode string) error {
	return auth.NewAuthService(s.db).VerifyPrivilegedStepUp(userID, password, mfaCode)
}

func DatasetNames() []string {
	return []string{"members", "family_members", "membership_fees", "sales", "transactions", "payments", "requests", "stock", "expenses", "fines", "letters", "committee", "audit_logs"}
}

func (s *Service) WriteDatasetCSV(writer io.Writer, dataset string, filter ExportFilter) error {
	definition, ok := datasets[dataset]
	if !ok {
		return fmt.Errorf("unsupported export dataset")
	}
	query, args, err := buildDatasetQuery(definition, filter)
	if err != nil {
		return err
	}
	rows, err := s.db.Raw(query, args...).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	return writeSQLRowsCSV(writer, rows, config.AppConfig.ExportMaxRows)
}

func (s *Service) CreateAllDataExport(filter ExportFilter, passphrase string) (string, string, error) {
	if !s.exportMu.TryLock() {
		return "", "", errors.New("another full data export is already running")
	}
	defer s.exportMu.Unlock()
	return s.createEncryptedArchive("data-export", passphrase, func(zipWriter *zip.Writer, hashes map[string]string) error {
		for _, name := range DatasetNames() {
			definition := datasets[name]
			entry, err := zipWriter.CreateHeader(&zip.FileHeader{Name: definition.Filename, Method: zip.Deflate})
			if err != nil {
				return err
			}
			hash := sha256.New()
			multi := io.MultiWriter(entry, hash)
			if err := s.WriteDatasetCSV(multi, name, filter); err != nil {
				return fmt.Errorf("export %s: %w", name, err)
			}
			hashes[definition.Filename] = hex.EncodeToString(hash.Sum(nil))
		}
		return nil
	})
}

func (s *Service) CreateDatabaseBackup(passphrase string) (string, string, error) {
	if !s.backupMu.TryLock() {
		return "", "", errors.New("another backup is already running")
	}
	defer s.backupMu.Unlock()
	workDir, err := os.MkdirTemp(config.AppConfig.BackupTempDir, "bansamiti-db-backup-")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(workDir)
	dumpPath := filepath.Join(workDir, "database.dump")
	if err := s.runPGDump(dumpPath); err != nil {
		return "", "", err
	}
	filename := "bansamiti-database-" + time.Now().UTC().Format("20060102-150405") + ".bsmbackup"
	outputPath, err := reserveUniqueOutput(filename)
	if err != nil {
		return "", "", err
	}
	if err := backupcrypto.EncryptFile(dumpPath, outputPath, passphrase); err != nil {
		return "", "", err
	}
	return outputPath, filename, nil
}

func (s *Service) CreateFullBackup(passphrase string) (string, string, error) {
	if !s.backupMu.TryLock() {
		return "", "", errors.New("another backup is already running")
	}
	defer s.backupMu.Unlock()
	return s.createEncryptedArchive("full-backup", passphrase, func(zipWriter *zip.Writer, hashes map[string]string) error {
		workDir, err := os.MkdirTemp(config.AppConfig.BackupTempDir, "bansamiti-full-backup-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(workDir)
		dumpPath := filepath.Join(workDir, "database.dump")
		if err := s.runPGDump(dumpPath); err != nil {
			return err
		}
		if err := addFileToZip(zipWriter, dumpPath, "database/database.dump", hashes); err != nil {
			return err
		}
		return addDirectoryToZip(zipWriter, config.AppConfig.UploadDir, "uploads", hashes)
	})
}

func (s *Service) createEncryptedArchive(kind, passphrase string, populate func(*zip.Writer, map[string]string) error) (string, string, error) {
	if len(passphrase) < 16 {
		return "", "", errors.New("backup passphrase must contain at least 16 characters")
	}
	workDir, err := os.MkdirTemp(config.AppConfig.BackupTempDir, "bansamiti-archive-")
	if err != nil {
		return "", "", err
	}
	defer os.RemoveAll(workDir)
	zipPath := filepath.Join(workDir, kind+".zip")
	zipFile, err := os.OpenFile(zipPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", "", err
	}
	zipWriter := zip.NewWriter(zipFile)
	hashes := make(map[string]string)
	if err := populate(zipWriter, hashes); err != nil {
		_ = zipWriter.Close()
		_ = zipFile.Close()
		return "", "", err
	}
	manifest := backupManifest{
		FormatVersion: "1", Application: "Ban Samiti Management System", AppVersion: config.AppConfig.AppVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Database: config.AppConfig.DBName,
		IncludesFiles: kind == "full-backup", Files: hashes,
		RestoreNotes: []string{
			"Decrypt with: go run ./cmd/backupctl decrypt -in <backup.bsmbackup> -out <archive.zip>",
			"Restore database.dump with pg_restore after reviewing the target database and roles.",
			"Never restore an untrusted or unverified archive directly into production.",
		},
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	entry, err := zipWriter.CreateHeader(&zip.FileHeader{Name: "manifest.json", Method: zip.Deflate})
	if err != nil {
		_ = zipWriter.Close()
		_ = zipFile.Close()
		return "", "", err
	}
	if _, err := entry.Write(manifestBytes); err != nil {
		_ = zipWriter.Close()
		_ = zipFile.Close()
		return "", "", err
	}
	if err := zipWriter.Close(); err != nil {
		_ = zipFile.Close()
		return "", "", err
	}
	if err := zipFile.Close(); err != nil {
		return "", "", err
	}
	filename := fmt.Sprintf("bansamiti-%s-%s.bsmbackup", kind, time.Now().UTC().Format("20060102-150405"))
	outputPath, err := reserveUniqueOutput(filename)
	if err != nil {
		return "", "", err
	}
	if err := backupcrypto.EncryptFile(zipPath, outputPath, passphrase); err != nil {
		return "", "", err
	}
	return outputPath, filename, nil
}

func (s *Service) runPGDump(outputPath string) error {
	timeout := time.Duration(config.AppConfig.BackupTimeoutMinutes) * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := []string{
		"--format=custom", "--compress=6", "--no-owner", "--no-acl",
		"--host=" + config.AppConfig.DBHost,
		"--port=" + config.AppConfig.DBPort,
		"--username=" + config.AppConfig.DBUser,
		"--file=" + outputPath,
		config.AppConfig.DBName,
	}
	command := exec.CommandContext(ctx, config.AppConfig.PGDumpPath, args...)
	command.Env = append(os.Environ(), "PGPASSWORD="+config.AppConfig.DBPassword)
	var stderr strings.Builder
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.New("database backup timed out")
		}
		return fmt.Errorf("pg_dump failed: %s", sanitizeToolError(stderr.String()))
	}
	info, err := os.Stat(outputPath)
	if err != nil || info.Size() == 0 {
		return errors.New("pg_dump did not produce a valid backup file")
	}
	return nil
}

func buildDatasetQuery(definition datasetDefinition, filter ExportFilter) (string, []interface{}, error) {
	query := definition.SQL
	args := make([]interface{}, 0, 3)
	if filter.FiscalYearID > 0 {
		if definition.FiscalColumn == "" {
			return "", nil, fmt.Errorf("this dataset does not support fiscal-year filtering")
		}
		query += " AND " + definition.FiscalColumn + " = ?"
		args = append(args, filter.FiscalYearID)
	}
	if filter.FromDate != "" {
		if definition.DateColumn == "" {
			return "", nil, fmt.Errorf("this dataset does not support date filtering")
		}
		from, err := time.Parse("2006-01-02", filter.FromDate)
		if err != nil {
			return "", nil, fmt.Errorf("invalid from_date")
		}
		query += " AND " + definition.DateColumn + " >= ?"
		args = append(args, from)
	}
	if filter.ToDate != "" {
		if definition.DateColumn == "" {
			return "", nil, fmt.Errorf("this dataset does not support date filtering")
		}
		to, err := time.Parse("2006-01-02", filter.ToDate)
		if err != nil {
			return "", nil, fmt.Errorf("invalid to_date")
		}
		query += " AND " + definition.DateColumn + " < ?"
		args = append(args, to.Add(24*time.Hour))
	}
	query += " ORDER BY 1"
	return query, args, nil
}

func writeSQLRowsCSV(writer io.Writer, rows *sql.Rows, maxRows int) error {
	if _, err := writer.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	csvWriter := csv.NewWriter(writer)
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if err := csvWriter.Write(columns); err != nil {
		return err
	}
	rowCount := 0
	for rows.Next() {
		rowCount++
		if rowCount > maxRows {
			return fmt.Errorf("export exceeds configured maximum of %d rows", maxRows)
		}
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return err
		}
		record := make([]string, len(columns))
		for index, value := range values {
			record[index] = csvSafe(formatSQLValue(value))
		}
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func formatSQLValue(value interface{}) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	case time.Time:
		return typed.UTC().Format(time.RFC3339)
	case bool:
		return strconv.FormatBool(typed)
	default:
		return fmt.Sprint(typed)
	}
}

func csvSafe(value string) string {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

func addDirectoryToZip(writer *zip.Writer, root, prefix string, hashes map[string]string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if _, err := os.Stat(absoluteRoot); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return filepath.WalkDir(absoluteRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(absoluteRoot, path)
		if err != nil || strings.HasPrefix(relative, "..") {
			return errors.New("unsafe upload path")
		}
		return addFileToZip(writer, path, filepath.ToSlash(filepath.Join(prefix, relative)), hashes)
	})
}

func addFileToZip(writer *zip.Writer, path, archiveName string, hashes map[string]string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("backup input is not a regular file: %s", path)
	}
	header := &zip.FileHeader{Name: archiveName, Method: zip.Deflate}
	header.SetMode(0o600)
	header.SetModTime(info.ModTime())
	entry, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(entry, hash), file); err != nil {
		return err
	}
	hashes[archiveName] = hex.EncodeToString(hash.Sum(nil))
	return nil
}

func reserveUniqueOutput(filename string) (string, error) {
	extension := filepath.Ext(filename)
	prefix := strings.TrimSuffix(filename, extension) + "-"
	file, err := os.CreateTemp(os.TempDir(), prefix+"*"+extension)
	if err != nil {
		return "", err
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
}

func sanitizeToolError(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, config.AppConfig.DBPassword, "[redacted]")
	if len(value) > 500 {
		value = value[:500]
	}
	if value == "" {
		return "backup utility returned an error"
	}
	return value
}
