package utils

import (
	"errors"
	"strings"

	"forest-management/config"
)

type SMSService interface {
	SendSMS(phone string, message string) error
}

type DisabledSMS struct{}

func (DisabledSMS) SendSMS(string, string) error {
	return errors.New("SMS delivery is not configured")
}

// SparrowSMS intentionally refuses to claim delivery until the official API
// integration is configured and tested. It never prints credentials/messages.
type SparrowSMS struct {
	APIKey   string
	SenderID string
}

func (s *SparrowSMS) SendSMS(phone string, message string) error {
	if strings.TrimSpace(s.APIKey) == "" || strings.TrimSpace(phone) == "" || strings.TrimSpace(message) == "" {
		return errors.New("SMS delivery is not configured")
	}
	return errors.New("Sparrow SMS transport is not implemented; temporary credentials were not sent")
}

func GetSMSService() SMSService {
	cfg := config.AppConfig
	if cfg == nil || strings.ToLower(cfg.SMSProvider) == "disabled" {
		return DisabledSMS{}
	}
	if strings.ToLower(cfg.SMSProvider) == "sparrow" {
		return &SparrowSMS{APIKey: cfg.SMSAPIKey, SenderID: cfg.SMSSenderID}
	}
	return DisabledSMS{}
}
