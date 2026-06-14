package fileutil

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forest-management/config"

	"github.com/google/uuid"
)

const DefaultMaxBytes int64 = 10 * 1024 * 1024

type Policy struct {
	MaxBytes int64
	Allowed  map[string]string // detected MIME -> canonical extension
}

type SavedFile struct {
	StoredName string
	Path       string
	URL        string
	MimeType   string
	Size       int64
	SHA256     string
}

var ImagePolicy = Policy{MaxBytes: 8 * 1024 * 1024, Allowed: map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}}

var EvidencePolicy = Policy{MaxBytes: 10 * 1024 * 1024, Allowed: map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}}

// Save writes an upload outside the public web root using a random canonical
// filename. It does not trust the browser-provided extension or MIME type.
func Save(reader io.Reader, folder, prefix string, policy Policy) (*SavedFile, error) {
	if reader == nil {
		return nil, errors.New("file is required")
	}
	folder = filepath.Base(strings.TrimSpace(folder))
	if folder == "." || folder == "" || strings.Contains(folder, "..") {
		return nil, errors.New("invalid upload folder")
	}
	if policy.MaxBytes <= 0 {
		policy.MaxBytes = DefaultMaxBytes
	}
	if len(policy.Allowed) == 0 {
		return nil, errors.New("upload policy has no allowed file types")
	}

	root, err := filepath.Abs(config.AppConfig.UploadDir)
	if err != nil {
		return nil, errors.New("invalid upload directory")
	}
	directory := filepath.Join(root, folder)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}

	temporary, err := os.CreateTemp(directory, ".upload-*")
	if err != nil {
		return nil, fmt.Errorf("create upload file: %w", err)
	}
	temporaryPath := temporary.Name()
	_ = temporary.Chmod(0o600)
	removeTemporary := true
	defer func() {
		_ = temporary.Close()
		if removeTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()

	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(temporary, hasher), io.LimitReader(reader, policy.MaxBytes+1))
	if copyErr != nil {
		return nil, fmt.Errorf("save upload: %w", copyErr)
	}
	if written <= 0 {
		return nil, errors.New("empty files are not allowed")
	}
	if written > policy.MaxBytes {
		return nil, fmt.Errorf("file exceeds the %d MB limit", policy.MaxBytes/(1024*1024))
	}
	if err := temporary.Sync(); err != nil {
		return nil, fmt.Errorf("flush upload: %w", err)
	}
	if _, err := temporary.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	header := make([]byte, 512)
	n, err := temporary.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	mimeType := strings.TrimSpace(strings.Split(http.DetectContentType(header[:n]), ";")[0])
	extension, allowed := policy.Allowed[mimeType]
	if !allowed {
		return nil, fmt.Errorf("detected file type %s is not allowed", mimeType)
	}
	if err := temporary.Close(); err != nil {
		return nil, err
	}

	if err := scanWithClamAV(temporaryPath); err != nil {
		return nil, err
	}

	prefix = sanitizePrefix(prefix)
	storedName := prefix + uuid.NewString() + extension
	finalPath := filepath.Join(directory, storedName)
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		return nil, fmt.Errorf("finalize upload: %w", err)
	}
	removeTemporary = false
	if err := os.Chmod(finalPath, 0o640); err != nil {
		_ = os.Remove(finalPath)
		return nil, fmt.Errorf("secure upload permissions: %w", err)
	}

	return &SavedFile{
		StoredName: storedName,
		Path:       finalPath,
		URL:        fmt.Sprintf("/uploads/%s/%s", folder, storedName),
		MimeType:   mimeType,
		Size:       written,
		SHA256:     hex.EncodeToString(hasher.Sum(nil)),
	}, nil
}

func RemoveURL(fileURL string) error {
	if strings.TrimSpace(fileURL) == "" {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(fileURL, "/uploads/"), "/")
	if len(parts) != 2 || filepath.Base(parts[0]) != parts[0] || filepath.Base(parts[1]) != parts[1] {
		return errors.New("invalid stored file URL")
	}
	root, err := filepath.Abs(config.AppConfig.UploadDir)
	if err != nil {
		return err
	}
	candidate := filepath.Join(root, parts[0], parts[1])
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return errors.New("unsafe stored file path")
	}
	if err := os.Remove(candidate); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func sanitizePrefix(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			builder.WriteRune(char)
		}
	}
	if builder.Len() == 0 {
		return "file_"
	}
	return builder.String() + "_"
}

func scanWithClamAV(path string) error {
	address := strings.TrimSpace(config.AppConfig.ClamAVAddr)
	if address == "" {
		if config.AppConfig.RequireAntivirus {
			return errors.New("antivirus scanning is required but CLAMAV_ADDR is not configured")
		}
		return nil
	}
	connection, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		if config.AppConfig.RequireAntivirus {
			return fmt.Errorf("antivirus service unavailable: %w", err)
		}
		return nil
	}
	defer connection.Close()
	_ = connection.SetDeadline(time.Now().Add(30 * time.Second))
	if _, err := connection.Write([]byte("zINSTREAM\x00")); err != nil {
		return antivirusFailure(err)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	buffer := make([]byte, 32*1024)
	for {
		n, readErr := file.Read(buffer)
		if n > 0 {
			var size [4]byte
			binary.BigEndian.PutUint32(size[:], uint32(n))
			if _, err := connection.Write(size[:]); err != nil {
				return antivirusFailure(err)
			}
			if _, err := connection.Write(buffer[:n]); err != nil {
				return antivirusFailure(err)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	if _, err := connection.Write([]byte{0, 0, 0, 0}); err != nil {
		return antivirusFailure(err)
	}
	result, err := bufio.NewReader(connection).ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return antivirusFailure(err)
	}
	result = strings.TrimSpace(strings.TrimRight(result, "\x00"))
	if strings.Contains(result, "FOUND") {
		return errors.New("upload rejected: malware was detected")
	}
	if !strings.Contains(result, "OK") {
		return antivirusFailure(fmt.Errorf("unexpected scanner response"))
	}
	return nil
}

func antivirusFailure(err error) error {
	if config.AppConfig.RequireAntivirus {
		return fmt.Errorf("antivirus scan failed: %w", err)
	}
	return nil
}
