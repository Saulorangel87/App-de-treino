package feedback

import (
	"strings"
	"testing"
	"time"
)

func TestDigestEscapesUserContent(t *testing.T) {
	entries := []DigestEntry{{
		DisplayName: "Ciclista <um>",
		Category:    CategoryExperience,
		Rating:      5,
		Message:     "Usei <strong>e gostei</strong> & recomendo.",
		CreatedAt:   time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC),
	}}
	htmlBody := DigestHTML(entries)
	if !strings.Contains(htmlBody, "Ciclista &lt;um&gt;") || !strings.Contains(htmlBody, "&lt;strong&gt;e gostei&lt;/strong&gt;") {
		t.Fatalf("digest did not escape user content: %s", htmlBody)
	}
	if strings.Contains(htmlBody, "<strong>e gostei</strong>") {
		t.Fatal("user content must not become HTML")
	}
}
