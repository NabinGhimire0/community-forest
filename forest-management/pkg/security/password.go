package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

var blockedPasswords = map[string]struct{}{
	"password": {}, "password123": {}, "admin123": {}, "123456789": {},
	"qwerty123": {}, "nepal123": {}, "bansamiti": {}, "letmein": {},
}

func ValidatePassword(password string) error {
	if len(password) < 12 {
		return fmt.Errorf("password must contain at least 12 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}
	if _, blocked := blockedPasswords[strings.ToLower(strings.TrimSpace(password))]; blocked {
		return fmt.Errorf("this password is too common")
	}
	var upper, lower, digit, symbol bool
	for _, value := range password {
		switch {
		case unicode.IsUpper(value):
			upper = true
		case unicode.IsLower(value):
			lower = true
		case unicode.IsDigit(value):
			digit = true
		case unicode.IsPunct(value) || unicode.IsSymbol(value):
			symbol = true
		}
	}
	if !upper || !lower || !digit || !symbol {
		return fmt.Errorf("password must include uppercase, lowercase, number, and symbol")
	}
	return nil
}

// GenerateStrongPassword returns a temporary password that satisfies the
// production password policy. It should be shown once and replaced at first login.
func GenerateStrongPassword(length int) (string, error) {
	if length < 16 {
		length = 16
	}
	groups := []string{
		"ABCDEFGHJKLMNPQRSTUVWXYZ",
		"abcdefghjkmnpqrstuvwxyz",
		"23456789",
		"!@#$%&*?",
	}
	all := strings.Join(groups, "")
	result := make([]byte, length)
	for index, group := range groups {
		value, err := randomCharacter(group)
		if err != nil {
			return "", err
		}
		result[index] = value
	}
	for index := len(groups); index < length; index++ {
		value, err := randomCharacter(all)
		if err != nil {
			return "", err
		}
		result[index] = value
	}
	// Fisher-Yates shuffle using crypto/rand.
	for index := len(result) - 1; index > 0; index-- {
		value, err := rand.Int(rand.Reader, big.NewInt(int64(index+1)))
		if err != nil {
			return "", err
		}
		swap := int(value.Int64())
		result[index], result[swap] = result[swap], result[index]
	}
	return string(result), nil
}

func randomCharacter(characters string) (byte, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
	if err != nil {
		return 0, err
	}
	return characters[value.Int64()], nil
}
