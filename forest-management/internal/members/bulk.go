package members

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"forest-management/internal/models"
	"forest-management/pkg/utils"

	"gorm.io/gorm"
)

type BulkImportService struct {
	db  *gorm.DB
	sms utils.SMSService
}

func NewBulkImportService(db *gorm.DB) *BulkImportService {
	return &BulkImportService{
		db:  db,
		sms: utils.GetSMSService(),
	}
}

type ImportResult struct {
	Row          int    `json:"row"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Status       string `json:"status"` // success, error, skipped
	Error        string `json:"error,omitempty"`
	MembershipNo string `json:"membership_no,omitempty"`
	Password     string `json:"password,omitempty"` // Only shown for success
}

type BulkImportResponse struct {
	TotalRows    int            `json:"total_rows"`
	SuccessCount int            `json:"success_count"`
	ErrorCount   int            `json:"error_count"`
	SkippedCount int            `json:"skipped_count"`
	Results      []ImportResult `json:"results"`
}

// ImportFromCSV reads a CSV file and imports members in bulk
//
// CSV Format (header row required):
// name,assistant_name,father_name,ward_no,tole,phone,joined_date,remarks
//
// Example:
// Ram Bahadur,Shyam Bahadur,Hari Bahadur,5,Nayabazar,9801234567,2024-01-15,Initial member
// Sita Devi,Gita Devi,Ram Kumar,3,Main Road,9807654321,,,
func (s *BulkImportService) ImportFromCSV(reader io.Reader) (*BulkImportResponse, error) {
	csvReader := csv.NewReader(reader)

	// Read header row
	headers, err := csvReader.Read()
	if err != nil {
		return nil, errors.New("failed to read CSV header")
	}

	// Validate required headers
	requiredHeaders := []string{"name", "assistant_name", "father_name", "ward_no", "tole", "phone"}
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[trim(strings.ToLower(h))] = i
	}

	for _, req := range requiredHeaders {
		if _, exists := headerMap[req]; !exists {
			return nil, fmt.Errorf("missing required column: %s", req)
		}
	}

	var results []ImportResult
	rowNum := 1 // Header is row 0
	successCount := 0
	errorCount := 0
	skippedCount := 0

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			results = append(results, ImportResult{
				Row:    rowNum,
				Status: "error",
				Error:  fmt.Sprintf("CSV parse error: %v", err),
			})
			errorCount++
			rowNum++
			continue
		}

		rowNum++
		result := s.importRow(row, headerMap)
		results = append(results, result)

		switch result.Status {
		case "success":
			successCount++
		case "error":
			errorCount++
		case "skipped":
			skippedCount++
		}
	}

	return &BulkImportResponse{
		TotalRows:    len(results),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		SkippedCount: skippedCount,
		Results:      results,
	}, nil
}

func (s *BulkImportService) importRow(row []string, headerMap map[string]int) ImportResult {
	getCol := func(key string) string {
		if idx, exists := headerMap[key]; exists && idx < len(row) {
			return trim(row[idx])
		}
		return ""
	}

	name := getCol("name")
	assistantName := getCol("assistant_name")
	fatherName := getCol("father_name")
	phone := getCol("phone")
	wardNoStr := getCol("ward_no")
	tole := getCol("tole")
	joinedDateStr := getCol("joined_date")
	remarks := getCol("remarks")

	// Validate required fields
	if name == "" {
		return ImportResult{Row: headerMap["name"] + 1, Name: name, Phone: phone, Status: "error", Error: "name is required"}
	}
	if phone == "" {
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "phone is required"}
	}
	if wardNoStr == "" {
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "ward_no is required"}
	}

	wardNo, err := strconv.Atoi(wardNoStr)
	if err != nil {
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "invalid ward_no"}
	}

	// Check if phone already exists
	var existingUser models.User
	if s.db.Where("phone = ?", phone).First(&existingUser).Error == nil {
		return ImportResult{Name: name, Phone: phone, Status: "skipped", Error: "phone number already registered"}
	}

	// Generate membership number
	tx := s.db.Begin()
	membershipNo, err := s.generateMembershipNo(tx)
	if err != nil {
		tx.Rollback()
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "failed to generate membership number"}
	}

	// Generate password
	plainPassword := generateRandomPassword(8)
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		tx.Rollback()
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "failed to hash password"}
	}

	// Create User
	user := models.User{
		Name:     name,
		Phone:    phone,
		Password: hashedPassword,
		Role:     "member",
		Status:   "active",
	}
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "failed to create user (phone may be duplicate)"}
	}

	// Parse joined date
	var joinedDate *time.Time
	if joinedDateStr != "" {
		t, err := time.Parse("2006-01-02", joinedDateStr)
		if err == nil {
			joinedDate = &t
		}
	}

	// Create Member
	var remarksPtr *string
	if remarks != "" {
		remarksPtr = &remarks
	}

	member := models.Member{
		UserID:        &user.ID,
		MembershipNo:  membershipNo,
		Name:          name,
		AssistantName: assistantName,
		FatherName:    fatherName,
		WardNo:        wardNo,
		Tole:          tole,
		Phone:         &phone,
		JoinedDate:    joinedDate,
		Status:        "active",
		Remarks:       remarksPtr,
	}

	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "failed to create member"}
	}

	if err := tx.Commit().Error; err != nil {
		return ImportResult{Name: name, Phone: phone, Status: "error", Error: "failed to commit transaction"}
	}

	// Send SMS (async)
	go s.sendCredentialsSMS(phone, plainPassword)

	return ImportResult{
		Name:         name,
		Phone:        phone,
		Status:       "success",
		MembershipNo: membershipNo,
		Password:     plainPassword,
	}
}

func (s *BulkImportService) generateMembershipNo(tx *gorm.DB) (string, error) {
	var lastMember models.Member
	err := tx.Order("id DESC").First(&lastMember).Error

	var nextNum int
	if err != nil {
		nextNum = 1
	} else {
		var lastNum int
		_, err := fmt.Sscanf(lastMember.MembershipNo, "MEM-%04d", &lastNum)
		if err != nil {
			var count int64
			tx.Model(&models.Member{}).Count(&count)
			nextNum = int(count) + 1
		} else {
			nextNum = lastNum + 1
		}
	}

	return fmt.Sprintf("MEM-%04d", nextNum), nil
}

func (s *BulkImportService) sendCredentialsSMS(phone, password string) {
	message := fmt.Sprintf(
		"Ban Samiti: Your account has been created.\nPhone: %s\nPassword: %s",
		phone, password,
	)
	if err := s.sms.SendSMS(phone, message); err != nil {
		fmt.Printf("⚠️ Failed to send SMS to %s: %v\n", phone, err)
	}
}

// GenerateCSVTemplate returns a CSV template for bulk import
func GenerateCSVTemplate() []byte {
	template := "name,assistant_name,father_name,ward_no,tole,phone,joined_date,remarks\n"
	template += "Ram Bahadur,Shyam Bahadur,Hari Bahadur,5,Nayabazar,9801234567,2024-01-15,Initial member\n"
	template += "Sita Devi,Gita Devi,Ram Kumar,3,Main Road,9807654321,,\n"
	return []byte(template)
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
