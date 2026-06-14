package security

import "testing"

func TestValidatePassword(t *testing.T) {
	valid := []string{"Forest@Secure2026", "Panchakanya#Admin42"}
	for _, password := range valid {
		if err := ValidatePassword(password); err != nil {
			t.Fatalf("valid password rejected: %v", err)
		}
	}
	invalid := []string{"short", "password123", "alllowercase123!", "NOLOWERCASE123!", "NoNumberAtAll!"}
	for _, password := range invalid {
		if err := ValidatePassword(password); err == nil {
			t.Fatalf("invalid password accepted: %q", password)
		}
	}
}
