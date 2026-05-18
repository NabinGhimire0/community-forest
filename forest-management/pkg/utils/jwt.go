package utils

import (
	"fmt"
	"strconv"
	"time"

	"forest-management/config"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a JWT token with user claims
// Claims are the "payload" data stored inside the token
func GenerateToken(userID uint, phone string, role string) (string, error) {
	// How long until token expires
	expiryHours, _ := strconv.Atoi(config.AppConfig.JWTExpiryHours)
	expiryTime := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	// Create claims (data stored in the token)
	claims := jwt.MapClaims{
		"user_id": userID,
		"phone":   phone,
		"role":    role,
		"exp":     expiryTime.Unix(), // Expiration time as Unix timestamp
		"iat":     time.Now().Unix(), // Issued at
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ParseToken extracts claims from a JWT token string
func ParseToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token with our secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Extract claims (payload)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}
