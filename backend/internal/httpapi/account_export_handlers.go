package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/i18n"
	"github.com/Saulorangel87/App-de-treino/backend/internal/legal"
	"github.com/Saulorangel87/App-de-treino/backend/internal/xlsx"
)

const spreadsheetContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// DataExporter builds the sheets of the athlete's data spreadsheet.
type DataExporter interface {
	ExportAccountSpreadsheet(ctx context.Context, userID string) ([]xlsx.Sheet, error)
}

// RouterOption configures optional collaborators without changing NewRouter's
// positional signature.
type RouterOption func(*Server)

func WithDataExporter(exporter DataExporter) RouterOption {
	return func(s *Server) { s.exporter = exporter }
}

// exportAccountData gives the signed-in athlete a spreadsheet with the essentials
// of their account and the training they completed (LGPD art. 18, II). It leaves
// out health data and credentials; the complete copy is available on request.
func (s *Server) exportAccountData(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if s.exporter == nil {
		writeError(w, http.StatusNotFound, "not_available", "A exportação de dados não está disponível.")
		return
	}
	sheets, err := s.exporter.ExportAccountSpreadsheet(r.Context(), user.ID)
	if err != nil || len(sheets) == 0 {
		slog.Default().Error("account export failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível exportar seus dados agora.")
		return
	}
	language := i18n.FromContext(r.Context())
	note := []any{
		"Observação",
		"Esta planilha traz o essencial. Para uma cópia completa dos seus dados, inclusive saúde e limitações, escreva para " + legal.ContactEmail + ".",
	}
	if language == i18n.English {
		note = []any{"Note", "This spreadsheet has the essentials. For a full copy of your data, including health data and limitations, write to " + legal.ContactEmail + "."}
	}
	sheets[0].Rows = append(sheets[0].Rows, note)
	translateSheets(language, sheets)
	var body bytes.Buffer
	if err := xlsx.Write(&body, sheets); err != nil {
		slog.Default().Error("account export encoding failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível exportar seus dados agora.")
		return
	}
	w.Header().Set("Content-Type", spreadsheetContentType)
	filename := "cadencia-dados-"
	if language == i18n.English {
		filename = "cadencia-data-"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+time.Now().UTC().Format("2006-01-02")+`.xlsx"`)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body.Bytes())
}

// translateSheets turns sheet names, headers and fixed labels into the request
// language. Workout names and objectives come from the engine and translate the
// same way; dates stay in the Brazilian format the header announces.
func translateSheets(language i18n.Language, sheets []xlsx.Sheet) {
	if language == i18n.Portuguese {
		return
	}
	for index := range sheets {
		sheet := &sheets[index]
		sheet.Name = i18n.T(language, sheet.Name)
		// The headers are shared package variables: translate a copy.
		header := make([]string, len(sheet.Header))
		for column, title := range sheet.Header {
			header[column] = i18n.T(language, title)
		}
		sheet.Header = header
		for _, row := range sheet.Rows {
			for column, cell := range row {
				if text, ok := cell.(string); ok {
					row[column] = i18n.T(language, text)
				}
			}
		}
	}
}
