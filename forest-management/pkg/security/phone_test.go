package security

import "testing"

func TestNormalizeNepalMobile(t *testing.T) {
	tests := map[string]string{
		"9800000000":        "9800000000",
		"+977 980-000-0000": "9800000000",
		"9779800000000":     "9800000000",
	}
	for input, expected := range tests {
		actual, err := NormalizeNepalMobile(input)
		if err != nil {
			t.Fatalf("NormalizeNepalMobile(%q): %v", input, err)
		}
		if actual != expected {
			t.Fatalf("NormalizeNepalMobile(%q) = %q; want %q", input, actual, expected)
		}
	}
}

func TestNormalizeNepalMobileRejectsInvalid(t *testing.T) {
	for _, input := range []string{"", "123", "98000abc00", "+1 555 555 5555"} {
		if _, err := NormalizeNepalMobile(input); err == nil {
			t.Fatalf("NormalizeNepalMobile(%q) unexpectedly succeeded", input)
		}
	}
}
