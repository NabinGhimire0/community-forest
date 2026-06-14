package security

import (
	"errors"
	"strings"
	"unicode"
)

// NormalizeNepalMobile canonicalizes a Nepal mobile number for storage and
// login. It accepts 98XXXXXXXX or +97798XXXXXXXX-style input, removes common
// visual separators, and returns the canonical 10-digit form.
func NormalizeNepalMobile(input string) (string, error) {
	var builder strings.Builder
	for _, char := range strings.TrimSpace(input) {
		switch {
		case unicode.IsDigit(char):
			builder.WriteRune(char)
		case char == '+' || char == ' ' || char == '-' || char == '(' || char == ')':
			// Presentation characters are intentionally discarded.
		default:
			return "", errors.New("phone number contains invalid characters")
		}
	}

	value := builder.String()
	if strings.HasPrefix(value, "977") && len(value) == 13 {
		value = value[3:]
	}
	if len(value) != 10 || value[0] != '9' {
		return "", errors.New("phone number must be a valid 10-digit Nepal mobile number")
	}
	return value, nil
}
