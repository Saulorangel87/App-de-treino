package repository

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only when CADENCIA_TEST_DATABASE_URL points at a migrated database.
func TestAccountExportCoversEveryUserOwnedTable(t *testing.T) {
	url := os.Getenv("CADENCIA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("CADENCIA_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	store := New(pool)

	// Every section query must run, including for a user that owns nothing.
	sections, err := store.ExportAccountData(ctx, "00000000-0000-4000-8000-000000000000")
	if err != nil {
		t.Fatalf("export for an unknown user failed: %v", err)
	}
	if got := string(sections["profile"]); got != "null" {
		t.Fatalf("profile for an unknown user is %s, want null", got)
	}

	// Tables that hang off a user or an athlete profile must be exported, or
	// explicitly listed as security material that never leaves the server.
	notExported := map[string]bool{"auth_sessions": true, "auth_email_tokens": true}
	exported := map[string]bool{}
	for _, section := range accountExportSections {
		for _, table := range tablesIn(section.query) {
			exported[table] = true
		}
	}
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT table_name FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name IN ('user_id', 'athlete_profile_id', 'workout_session_id', 'training_plan_id')`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		if !exported[table] && !notExported[table] {
			t.Errorf("table %q holds user-owned data but is not part of the account export", table)
		}
	}
}

func tablesIn(query string) []string {
	var tables []string
	fields := strings.Fields(query)
	for i, field := range fields {
		if (field == "FROM" || field == "JOIN") && i+1 < len(fields) {
			tables = append(tables, strings.Trim(fields[i+1], "(),"))
		}
	}
	return tables
}
