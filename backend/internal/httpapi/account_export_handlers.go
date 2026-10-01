package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

const accountExportFormatVersion = 1

// DataExporter returns everything the app stores about one user, keyed by section.
type DataExporter interface {
	ExportAccountData(ctx context.Context, userID string) (map[string]json.RawMessage, error)
}

// RouterOption configures optional collaborators without changing NewRouter's
// positional signature.
type RouterOption func(*Server)

func WithDataExporter(exporter DataExporter) RouterOption {
	return func(s *Server) { s.exporter = exporter }
}

// exportAccountData gives the signed-in athlete a copy of their own data (LGPD
// art. 18, II and V). It never includes credentials or session tokens.
func (s *Server) exportAccountData(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if s.exporter == nil {
		writeError(w, http.StatusNotFound, "not_available", "A exportação de dados não está disponível.")
		return
	}
	sections, err := s.exporter.ExportAccountData(r.Context(), user.ID)
	if err != nil {
		slog.Default().Error("account export failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível exportar seus dados agora.")
		return
	}
	exportedAt := time.Now().UTC()
	document := map[string]any{
		"format_version": accountExportFormatVersion,
		"exported_at":    exportedAt.Format(time.RFC3339),
		"notice":         "Cópia dos dados que o Cadência guarda sobre você. Senhas e códigos de sessão não fazem parte dela.",
	}
	for key, value := range sections {
		document[key] = value
	}
	body, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível exportar seus dados agora.")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="cadencia-dados-`+exportedAt.Format("2006-01-02")+`.json"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
