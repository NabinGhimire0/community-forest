package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // TOTP interoperability requires HMAC-SHA1 by default (RFC 6238).
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	totpStepSeconds = int64(30)
	totpDigits      = 6
)

func GenerateTOTPSecret() (string, error) {
	buffer := make([]byte, 20)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(buffer), "="), nil
}

func TOTPURI(secret, account, issuer string) string {
	label := url.PathEscape(issuer + ":" + account)
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", strconv.Itoa(totpDigits))
	values.Set("period", strconv.FormatInt(totpStepSeconds, 10))
	return "otpauth://totp/" + label + "?" + values.Encode()
}

func VerifyTOTP(secret, code string, now time.Time) bool {
	_, valid := MatchTOTP(secret, code, now)
	return valid
}

// MatchTOTP returns the exact 30-second counter that matched the submitted
// code. Persisting this value lets callers reject reuse of the same code.
func MatchTOTP(secret, code string, now time.Time) (int64, bool) {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return 0, false
	}
	for _, value := range code {
		if value < '0' || value > '9' {
			return 0, false
		}
	}
	counter := now.Unix() / totpStepSeconds
	for offset := int64(-1); offset <= 1; offset++ {
		candidate := counter + offset
		if ConstantTimeStringEqual(generateTOTP(secret, candidate), code) {
			return candidate, true
		}
	}
	return 0, false
}

func generateTOTP(secret string, counter int64) string {
	padding := strings.Repeat("=", (8-len(secret)%8)%8)
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret) + padding)
	if err != nil {
		return ""
	}
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, uint64(counter))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(message)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", value%1000000)
}

func GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, 0, count)
	for index := 0; index < count; index++ {
		buffer := make([]byte, 5)
		if _, err := rand.Read(buffer); err != nil {
			return nil, err
		}
		raw := strings.ToUpper(hex.EncodeToString(buffer))
		codes = append(codes, raw[:5]+"-"+raw[5:])
	}
	return codes, nil
}

func HashBackupCode(code string, key []byte) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}
