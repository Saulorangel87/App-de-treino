package legal

import (
	"os"
	"regexp"
	"testing"
)

func TestConstantsMatchTheFrontendTexts(t *testing.T) {
	source, err := os.ReadFile("../../../frontend/lib/legal.ts")
	if err != nil {
		t.Skipf("frontend sources are not available: %v", err)
	}
	for name, want := range map[string]string{
		"LEGAL_VERSION":       TermsVersion,
		"LEGAL_CONTACT_EMAIL": ContactEmail,
	} {
		match := regexp.MustCompile(name + ` = '([^']+)'`).FindSubmatch(source)
		if match == nil {
			t.Fatalf("%s was not found in frontend/lib/legal.ts", name)
		}
		if got := string(match[1]); got != want {
			t.Fatalf("frontend %s is %s but the backend has %s; update both together", name, got, want)
		}
	}
}
