package utils

import (
	"fmt"
	"forest-management/config"
)

// SMSService is an interface — you can swap providers without changing business logic
type SMSService interface {
	SendSMS(phone string, message string) error
}

// SparrowSMS implements SMSService for Nepal's Sparrow SMS
type SparrowSMS struct {
	APIKey   string
	SenderID string
}

func NewSparrowSMS(apiKey, senderID string) *SparrowSMS {
	return &SparrowSMS{APIKey: apiKey, SenderID: senderID}
}

func (s *SparrowSMS) SendSMS(phone string, message string) error {
	// Sparrow SMS API implementation
	// In production, use net/http to POST to their endpoint
	fmt.Printf("📱 SMS sent to %s: %s\n", phone, message)
	// Actual API call would look like:
	// resp, err := http.PostForm("https://api.sparrowsms.com/v2/sms/",
	//     url.Values{
	//         "token":    {s.APIKey},
	//         "from":     {s.SenderID},
	//         "to":       {phone},
	//         "text":     {message},
	//     })
	return nil
}

// GetSMSService returns the configured SMS provider
func GetSMSService() SMSService {
	cfg := config.AppConfig // import "forest-management/config"
	switch cfg.SMSProvider {
	case "sparrow":
		return NewSparrowSMS(cfg.SMSAPIKey, cfg.SMSSenderID)
	// Add more providers:
	// case "twilio":
	//     return NewTwilioSMS(cfg.TwilioSID, cfg.TwilioToken, cfg.TwilioFrom)
	default:
		return NewSparrowSMS(cfg.SMSAPIKey, cfg.SMSSenderID)
	}
}
