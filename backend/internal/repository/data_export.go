package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// accountExportSections lists everything the app stores about one athlete. Each
// query returns a single jsonb value scoped to the user in $1. Credentials and
// session tokens are left out on purpose: they are security material, not the
// athlete's data. Any new user-owned table must be added here and to the
// account deletion fixture; the integration test fails if one is missing. It feeds
// cmd/account-export, which answers requests for a complete copy (LGPD art. 18).
var accountExportSections = []struct {
	key   string
	query string
}{
	{"account", `
		SELECT COALESCE((SELECT jsonb_build_object('id', u.id, 'email', u.email, 'display_name', u.display_name,
			'created_at', u.created_at, 'email_verified_at', u.email_verified_at)
			FROM users u WHERE u.id = $1), 'null'::jsonb)`},
	{"profile", `
		SELECT COALESCE((SELECT to_jsonb(ap) - 'user_id' FROM athlete_profiles ap WHERE ap.user_id = $1), 'null'::jsonb)`},
	{"goals", exportByProfile("goals", "g", "g.priority, g.created_at")},
	{"availability", exportByProfile("availability", "g", "g.weekday")},
	{"limitations", exportByProfile("injuries_or_limitations", "g", "g.created_at")},
	{"recovery_checkins", exportByProfile("recovery_data", "g", "g.recorded_on")},
	{"recovery_self_reports", exportByProfile("recovery_self_reports", "g", "g.reported_on")},
	{"cycling_assessments", exportByProfile("cycling_assessments", "g", "g.created_at")},
	{"training_plans", exportByProfile("training_plans", "g", "g.starts_on")},
	{"workouts", `
		SELECT COALESCE(jsonb_agg(to_jsonb(w) ORDER BY w.scheduled_on, w.created_at), '[]'::jsonb)
		FROM workouts w
		JOIN training_plans tp ON tp.id = w.training_plan_id
		JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1`},
	{"workout_sessions", exportByProfile("workout_sessions", "g", "g.created_at")},
	{"workout_feedback", `
		SELECT COALESCE(jsonb_agg(to_jsonb(f) ORDER BY f.created_at), '[]'::jsonb)
		FROM feedback f
		JOIN workout_sessions ws ON ws.id = f.workout_session_id
		JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
		WHERE ap.user_id = $1`},
	{"imported_activities", `
		SELECT COALESCE(jsonb_agg(to_jsonb(ia) - 'user_id' - 'file_hash' ORDER BY ia.started_at), '[]'::jsonb)
		FROM imported_activities ia WHERE ia.user_id = $1`},
	{"legal_acceptances", `
		SELECT COALESCE(jsonb_agg(to_jsonb(la) - 'user_id' ORDER BY la.accepted_at), '[]'::jsonb)
		FROM legal_acceptances la WHERE la.user_id = $1`},
	{"feedback_messages", `
		SELECT COALESCE(jsonb_agg(to_jsonb(uf) - 'user_id' ORDER BY uf.created_at), '[]'::jsonb)
		FROM user_feedback uf WHERE uf.user_id = $1`},
}

func exportByProfile(table, alias, order string) string {
	return fmt.Sprintf(`
		SELECT COALESCE(jsonb_agg(to_jsonb(%[2]s) - 'athlete_profile_id' ORDER BY %[3]s), '[]'::jsonb)
		FROM %[1]s %[2]s
		JOIN athlete_profiles ap ON ap.id = %[2]s.athlete_profile_id
		WHERE ap.user_id = $1`, table, alias, order)
}

// ExportAccountData returns every record the app holds about the user, keyed by
// section, for the athlete's right of access (LGPD art. 18). It reads inside one
// repeatable-read transaction so the sections are consistent with each other.
func (s *Store) ExportAccountData(ctx context.Context, userID string) (map[string]json.RawMessage, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	result := make(map[string]json.RawMessage, len(accountExportSections))
	for _, section := range accountExportSections {
		var raw []byte
		if err := tx.QueryRow(ctx, section.query, userID).Scan(&raw); err != nil {
			return nil, fmt.Errorf("export %s: %w", section.key, err)
		}
		result[section.key] = json.RawMessage(raw)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}
