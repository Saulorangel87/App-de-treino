// Package legal holds the version of the Terms of Use and Privacy Policy that
// users accept, and the contact for privacy requests. They must match
// LEGAL_VERSION and LEGAL_CONTACT_EMAIL in frontend/lib/legal.ts (a test enforces
// it); bump the version in both when either text changes in a way users must
// accept again.
package legal

const (
	TermsVersion = "2026-10-01"
	ContactEmail = "sauloleonardo1987@gmail.com"
)
