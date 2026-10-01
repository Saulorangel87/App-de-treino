package repository

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only when CADENCIA_TEST_DATABASE_URL points at a migrated database. It
// creates a user with health data and one completed workout, checks the
// spreadsheet, and deletes the user (everything cascades).
func TestAccountSpreadsheetHasTheEssentialsAndNoHealthData(t *testing.T) {
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

	email := fmt.Sprintf("spreadsheet-%d@example.invalid", time.Now().UnixNano())
	var userID, profileID, planID, workoutID string
	if err := pool.QueryRow(ctx, `INSERT INTO users (email, password_hash, display_name) VALUES ($1, 'segredo-do-hash', 'Atleta de Teste') RETURNING id::text`, email).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	defer pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, query, args...); err != nil {
			t.Fatalf("%v\n%s", err, query)
		}
	}
	if err := pool.QueryRow(ctx, `INSERT INTO athlete_profiles (user_id, experience_level, weight_kg, body_fat_percent) VALUES ($1, 'intermediate', 83.4, 21.7) RETURNING id::text`, userID).Scan(&profileID); err != nil {
		t.Fatal(err)
	}
	mustExec(`INSERT INTO injuries_or_limitations (athlete_profile_id, kind, description) VALUES ($1, 'pain', 'dor no joelho esquerdo')`, profileID)
	mustExec(`INSERT INTO recovery_data (athlete_profile_id, recorded_on, sleep_quality, stress_level, fatigue_level, notes) VALUES ($1, DATE '2099-01-01', 4, 2, 2, 'anotação de saúde')`, profileID)
	mustExec(`INSERT INTO legal_acceptances (user_id, terms_version) VALUES ($1, '2026-10-01')`, userID)
	if err := pool.QueryRow(ctx, `INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES ($1, DATE '2099-01-01', DATE '2099-01-28', 'active') RETURNING id::text`, profileID).Scan(&planID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, status) VALUES ($1, DATE '2099-01-01', 'Giro de teste', 'Base', 60, 4, '{}'::jsonb, 'completed') RETURNING id::text`, planID).Scan(&workoutID); err != nil {
		t.Fatal(err)
	}
	mustExec(`INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, distance_km, average_power_watts, average_heart_rate, status) VALUES ($1, $2, now() - interval '70 minutes', now(), 65, 5, 32.5, 180, 141, 'completed')`, workoutID, profileID)
	mustExec(`INSERT INTO imported_activities (user_id, workout_id, source, file_hash, started_at, moving_seconds, distance_km) VALUES ($1, $2, 'fit', repeat('ab', 32), now(), 3900, 33.1)`, userID, workoutID)

	sheets, err := New(pool).ExportAccountSpreadsheet(ctx, userID)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if len(sheets) != 3 {
		t.Fatalf("got %d sheets, want 3", len(sheets))
	}

	var dump strings.Builder
	for _, sheet := range sheets {
		dump.WriteString(strings.Join(sheet.Header, "|") + "\n")
		for _, row := range sheet.Rows {
			if len(row) != len(sheet.Header) {
				t.Fatalf("sheet %q: row has %d cells, header has %d", sheet.Name, len(row), len(sheet.Header))
			}
			for _, cell := range row {
				dump.WriteString(fmt.Sprint(cell) + "|")
			}
			dump.WriteString("\n")
		}
	}
	text := dump.String()
	for _, want := range []string{email, "Atleta de Teste", "Intermediário", "2026-09-30", "Giro de teste", "32.5", "180", "141", "33.1", "65"} {
		if !strings.Contains(text, want) {
			t.Errorf("the spreadsheet is missing %q:\n%s", want, text)
		}
	}
	for _, forbidden := range []string{"joelho", "anotação de saúde", "83.4", "21.7", "segredo-do-hash", strings.Repeat("ab", 32)} {
		if strings.Contains(text, forbidden) {
			t.Errorf("the spreadsheet must not contain %q:\n%s", forbidden, text)
		}
	}
	if got := len(sheets[1].Rows); got != 1 {
		t.Errorf("got %d completed workouts, want 1", got)
	}
}
