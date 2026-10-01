package repository

import "context"

// RecordLegalAcceptance stores that the user accepted the terms version. Accepting
// the same version again keeps the original timestamp.
func (s *Store) RecordLegalAcceptance(ctx context.Context, userID, version string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO legal_acceptances (user_id, terms_version) VALUES ($1, $2)
		ON CONFLICT (user_id, terms_version) DO NOTHING`, userID, version)
	return err
}

func (s *Store) HasAcceptedLegal(ctx context.Context, userID, version string) (bool, error) {
	var accepted bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM legal_acceptances WHERE user_id = $1 AND terms_version = $2)`,
		userID, version).Scan(&accepted)
	return accepted, err
}
