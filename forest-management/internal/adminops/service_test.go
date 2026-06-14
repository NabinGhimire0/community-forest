package adminops

import "testing"

func TestCSVSafePreventsSpreadsheetFormulaInjection(t *testing.T) {
	cases := map[string]string{
		"=SUM(1,2)": "'=SUM(1,2)",
		" +cmd":     "' +cmd",
		"-10":       "'-10",
		"@IMPORT":   "'@IMPORT",
		"normal":    "normal",
		"2080/81":   "2080/81",
	}
	for input, expected := range cases {
		if actual := csvSafe(input); actual != expected {
			t.Fatalf("csvSafe(%q) = %q; want %q", input, actual, expected)
		}
	}
}
