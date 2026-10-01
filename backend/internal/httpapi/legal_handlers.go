package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/legal"
)

// LegalAcceptanceStore records which version of the Terms of Use and Privacy
// Policy each user accepted.
type LegalAcceptanceStore interface {
	RecordLegalAcceptance(ctx context.Context, userID, version string) error
	HasAcceptedLegal(ctx context.Context, userID, version string) (bool, error)
}

func WithLegalAcceptances(store LegalAcceptanceStore) RouterOption {
	return func(s *Server) { s.legalAcceptances = store }
}

type legalStatus struct {
	TermsVersion string `json:"terms_version"`
	Accepted     bool   `json:"accepted"`
}

// legalStatusFor reports whether the user accepted the current texts. Without a
// store (tests, tools) nothing is required.
func (s *Server) legalStatusFor(ctx context.Context, userID string) (legalStatus, error) {
	status := legalStatus{TermsVersion: legal.TermsVersion, Accepted: true}
	if s.legalAcceptances == nil {
		return status, nil
	}
	accepted, err := s.legalAcceptances.HasAcceptedLegal(ctx, userID, legal.TermsVersion)
	status.Accepted = accepted
	return status, err
}

type acceptTermsRequest struct {
	TermsVersion string `json:"terms_version"`
}

// acceptTerms records that a signed-in user accepted the current texts, for
// accounts created before acceptance was recorded and for new versions.
func (s *Server) acceptTerms(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if s.legalAcceptances == nil {
		writeError(w, http.StatusNotFound, "not_available", "O aceite dos termos não está disponível.")
		return
	}
	var input acceptTermsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.TermsVersion != legal.TermsVersion {
		writeError(w, http.StatusConflict, "terms_outdated", "Os termos foram atualizados. Recarregue a página e leia a nova versão.")
		return
	}
	if err := s.legalAcceptances.RecordLegalAcceptance(r.Context(), user.ID, legal.TermsVersion); err != nil {
		slog.Default().Error("recording terms acceptance failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível registrar o aceite agora.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
