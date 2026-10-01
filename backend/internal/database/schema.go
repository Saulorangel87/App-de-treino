package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequiredMigrations lists every migration this binary expects to find in
// cadencia_schema_migrations. TestRequiredMigrationsMatchFiles keeps it in sync
// with database/migrations, so adding a migration without updating this list
// fails the test suite.
var RequiredMigrations = []string{
	"000001_initial_schema.up.sql",
	"000002_auth_sessions.up.sql",
	"000003_single_active_plan.up.sql",
	"000004_workout_session_lifecycle.up.sql",
	"000005_adaptive_training_feedback.up.sql",
	"000006_training_plan_completion.up.sql",
	"000007_scientific_sources.up.sql",
	"000008_cycling_context.up.sql",
	"000009_submaximal_assessments.up.sql",
	"000010_interval_evidence.up.sql",
	"000011_workout_session_metrics.up.sql",
	"000012_email_verification_and_password_reset.up.sql",
	"000013_user_feedback.up.sql",
	"000014_feedback_digest.up.sql",
	"000015_cycling_catalog_evidence.up.sql",
	"000016_xco_catalog_evidence.up.sql",
	"000017_event_taper_evidence.up.sql",
	"000018_road_vo2_catalog_evidence.up.sql",
	"000019_short_intervals_evidence.up.sql",
	"000020_completion_context.up.sql",
	"000021_post_workout_context.up.sql",
	"000022_limitation_context.up.sql",
	"000023_feedback_context.up.sql",
	"000024_equipment_feedback.up.sql",
	"000025_limitation_safety_signals.up.sql",
	"000026_scientific_source_metadata.up.sql",
	"000027_adaptation_integrity_gate.up.sql",
	"000028_post_event_recovery_evidence.up.sql",
	"000029_average_cadence_metric.up.sql",
	"000030_profile_safety_context.up.sql",
	"000031_imported_activities.up.sql",
	"000032_recovery_self_reports.up.sql",
	"000033_fix_source_metadata.up.sql",
	"000034_audit_remaining_sources.up.sql",
	"000035_legal_acceptances.up.sql",
	"000036_task_mode.up.sql",
}

// ErrSchemaBehind reports that the database is missing migrations this binary needs.
var ErrSchemaBehind = errors.New("database schema is behind the application")

// MissingMigrations returns the required migrations that are not recorded in
// cadencia_schema_migrations. A missing table means no migration was recorded.
func MissingMigrations(ctx context.Context, query func(context.Context, string, ...any) (pgx.Rows, error)) ([]string, error) {
	rows, err := query(ctx, `SELECT filename FROM cadencia_schema_migrations`)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return append([]string(nil), RequiredMigrations...), nil
		}
		return nil, err
	}
	defer rows.Close()
	applied := make(map[string]struct{}, len(RequiredMigrations))
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var missing []string
	for _, name := range RequiredMigrations {
		if _, ok := applied[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing, nil
}

// SchemaChecker verifies the migration state of a pool.
type SchemaChecker struct{ pool *pgxpool.Pool }

func NewSchemaChecker(pool *pgxpool.Pool) *SchemaChecker { return &SchemaChecker{pool: pool} }

// Check returns ErrSchemaBehind, wrapped with the missing migration names, when
// the database does not have every required migration.
func (c *SchemaChecker) Check(ctx context.Context) error {
	missing, err := MissingMigrations(ctx, c.pool.Query)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %s", ErrSchemaBehind, strings.Join(missing, ", "))
	}
	return nil
}
