// Command account-export writes the complete copy of one athlete's data, as JSON,
// to the standard output. It answers a request for access (LGPD art. 18) that the
// spreadsheet in Settings does not cover, because that one leaves out health data.
//
// It only reads the database and is meant to be run by the product owner on the
// server, never exposed by the API:
//
//	cadencia-account-export --email atleta@exemplo.com > dados.json
//
// The JSON holds health data. Confirm the requester owns the account before
// sending it, protect the file with a password and delete it after sending.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/database"
	"github.com/Saulorangel87/App-de-treino/backend/internal/repository"
	"github.com/jackc/pgx/v5"
)

const formatVersion = 1

// source is what the command needs from the database.
type source interface {
	UserByEmail(ctx context.Context, email string) (auth.User, error)
	ExportAccountData(ctx context.Context, userID string) (map[string]json.RawMessage, error)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		logger.Error("DATABASE_URL is not set")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, url, database.Options{MaxConns: 1, MinConns: 0})
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := run(ctx, os.Args[1:], repository.New(db), os.Stdout, logger, time.Now); err != nil {
		logger.Error("account export failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, store source, out io.Writer, logger *slog.Logger, now func() time.Time) error {
	flags := flag.NewFlagSet("account-export", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	address := flags.String("email", "", "e-mail address of the account")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("usage: cadencia-account-export --email <address>: %w", err)
	}
	email, err := normalizeEmail(*address)
	if err != nil {
		return err
	}

	user, err := store.UserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no account uses %s", email)
	}
	if err != nil {
		return err
	}
	sections, err := store.ExportAccountData(ctx, user.ID)
	if err != nil {
		return err
	}

	exportedAt := now().UTC()
	document := map[string]any{
		"format_version": formatVersion,
		"exported_at":    exportedAt.Format(time.RFC3339),
		"notice":         "Cópia completa dos dados que o Cadência guarda sobre você, inclusive os de saúde. Senhas e códigos de sessão não fazem parte dela.",
	}
	for key, value := range sections {
		document[key] = value
	}
	body, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	if _, err := out.Write(append(body, '\n')); err != nil {
		return err
	}
	// Leaves a trace of when a copy was produced, without the content.
	logger.Info("complete account copy generated", "email", email, "at", exportedAt.Format(time.RFC3339), "sections", len(sections))
	return nil
}

func normalizeEmail(address string) (string, error) {
	address = strings.ToLower(strings.TrimSpace(address))
	if address == "" {
		return "", errors.New("usage: cadencia-account-export --email <address>")
	}
	parsed, err := mail.ParseAddress(address)
	if err != nil || parsed.Address != address {
		return "", fmt.Errorf("%q is not a valid e-mail address", address)
	}
	return address, nil
}
