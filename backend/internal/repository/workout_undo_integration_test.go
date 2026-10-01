package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Runs only when CADENCIA_TEST_DATABASE_URL points at a migrated database.
type undoFixture struct {
	t         *testing.T
	pool      *pgxpool.Pool
	store     *Store
	userID    string
	profileID string
	ctx       context.Context
}

func newUndoFixture(t *testing.T) *undoFixture {
	t.Helper()
	url := os.Getenv("CADENCIA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("CADENCIA_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	f := &undoFixture{t: t, pool: pool, store: New(pool), ctx: ctx}
	email := fmt.Sprintf("undo-%d@example.invalid", time.Now().UnixNano())
	f.scan(&f.userID, `INSERT INTO users (email, password_hash, display_name) VALUES ($1, 'x', 'Undo') RETURNING id::text`, email)
	t.Cleanup(func() { _, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, f.userID) })
	f.scan(&f.profileID, `INSERT INTO athlete_profiles (user_id, experience_level) VALUES ($1, 'intermediate') RETURNING id::text`, f.userID)
	return f
}

func (f *undoFixture) scan(dest any, query string, args ...any) {
	f.t.Helper()
	if err := f.pool.QueryRow(f.ctx, query, args...).Scan(dest); err != nil {
		f.t.Fatalf("%v\n%s", err, query)
	}
}

func (f *undoFixture) exec(query string, args ...any) {
	f.t.Helper()
	if _, err := f.pool.Exec(f.ctx, query, args...); err != nil {
		f.t.Fatalf("%v\n%s", err, query)
	}
}

func (f *undoFixture) plan(status string) string {
	var id string
	f.scan(&id, `INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES ($1, CURRENT_DATE - 5, CURRENT_DATE + 20, $2) RETURNING id::text`, f.profileID, status)
	return id
}

// workout inserts a workout dayOffset days from today, with a base rule in its explanation.
func (f *undoFixture) workout(planID string, dayOffset int, status string) string {
	var id string
	f.scan(&id, `
		INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
		VALUES ($1, CURRENT_DATE + $2::int, 'Treino', 'Base', 60, 6, '{}'::jsonb, '{"rules":["Regra base"]}'::jsonb, $3) RETURNING id::text`,
		planID, dayOffset, status)
	return id
}

// completeWithHardFeedback records a session that makes the trigger reduce the next two workouts.
func (f *undoFixture) completeWithHardFeedback(workoutID string) string {
	var sessionID string
	f.scan(&sessionID, `
		INSERT INTO workout_sessions (workout_id, athlete_profile_id, started_at, completed_at, duration_minutes, actual_rpe, status)
		VALUES ($1, $2, now() - interval '70 minutes', now(), 60, 9.5, 'completed') RETURNING id::text`, workoutID, f.profileID)
	f.exec(`UPDATE workouts SET status = 'completed' WHERE id = $1`, workoutID)
	f.exec(`INSERT INTO feedback (workout_session_id, completion_status, difficulty, pain_reported, fatigue_after) VALUES ($1, 'complete', 'very_hard', false, 5)`, sessionID)
	return sessionID
}

type workoutRow struct {
	status   string
	duration int
	rpe      float64
	adapted  bool
	rules    int
}

func (f *undoFixture) row(workoutID string) workoutRow {
	var r workoutRow
	f.scan(&r.status, `SELECT status FROM workouts WHERE id = $1`, workoutID)
	f.scan(&r.duration, `SELECT duration_minutes FROM workouts WHERE id = $1`, workoutID)
	f.scan(&r.rpe, `SELECT target_rpe::double precision FROM workouts WHERE id = $1`, workoutID)
	f.scan(&r.adapted, `SELECT explanation ? 'adaptation' FROM workouts WHERE id = $1`, workoutID)
	f.scan(&r.rules, `SELECT jsonb_array_length(explanation->'rules') FROM workouts WHERE id = $1`, workoutID)
	return r
}

func (f *undoFixture) count(query string, args ...any) int {
	var n int
	f.scan(&n, query, args...)
	return n
}

func TestUndoRestoresTheWorkoutAndRevertsTheAdaptationItCaused(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	done := f.workout(plan, -1, "planned")
	next1 := f.workout(plan, 1, "planned")
	next2 := f.workout(plan, 2, "planned")
	untouched := f.workout(plan, 3, "planned")
	session := f.completeWithHardFeedback(done)

	for _, id := range []string{next1, next2} {
		if r := f.row(id); r.status != "adapted" || r.duration >= 60 || !r.adapted || r.rules != 2 {
			t.Fatalf("precondition: the trigger should have reduced the next workouts, got %+v", r)
		}
	}
	if r := f.row(untouched); r.adapted || r.duration != 60 {
		t.Fatalf("precondition: only two workouts are adapted, got %+v", r)
	}

	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, done); err != nil {
		t.Fatalf("undo: %v", err)
	}

	for _, id := range []string{next1, next2} {
		if r := f.row(id); r.status != "planned" || r.duration != 60 || r.rpe != 6 || r.adapted || r.rules != 1 {
			t.Errorf("the adaptation was not reverted: %+v", r)
		}
	}
	if r := f.row(done); r.status != "planned" {
		t.Errorf("the undone workout is %q, want planned", r.status)
	}
	if n := f.count(`SELECT count(*) FROM workout_sessions WHERE id = $1`, session); n != 0 {
		t.Errorf("the session was not deleted")
	}
	if n := f.count(`SELECT count(*) FROM feedback WHERE workout_session_id = $1`, session); n != 0 {
		t.Errorf("the feedback was not deleted")
	}
	if r := f.row(untouched); r.status != "planned" || r.duration != 60 {
		t.Errorf("an unrelated workout changed: %+v", r)
	}

	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, done); !errors.Is(err, planning.ErrInvalidTransition) {
		t.Errorf("undoing a planned workout returned %v, want ErrInvalidTransition", err)
	}
}

func TestUndoOfASessionThatCausedNoAdaptationLeavesOtherAdaptationsAlone(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	first := f.workout(plan, -2, "planned")
	second := f.workout(plan, -1, "planned")
	target := f.workout(plan, 1, "planned")
	f.completeWithHardFeedback(first)
	// A recent protective signal stops the trigger from adapting again, so the
	// second session leaves no adaptation behind and the first one keeps ownership.
	f.completeWithHardFeedback(second)

	var source string
	f.scan(&source, `SELECT explanation #>> '{adaptation,source_workout_id}' FROM workouts WHERE id = $1`, target)
	if source != first {
		t.Fatalf("precondition: the first session should own the adaptation, got %s", source)
	}
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, second); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if r := f.row(target); !r.adapted || r.status != "adapted" || r.duration >= 60 {
		t.Errorf("undoing a session that adapted nothing must not revert another one's adaptation: %+v", r)
	}
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, first); err != nil {
		t.Fatalf("undo first: %v", err)
	}
	if r := f.row(target); r.adapted || r.status != "planned" || r.duration != 60 {
		t.Errorf("undoing the owner must revert it: %+v", r)
	}
}

func TestUndoReopensAPlanThatTheWorkoutHadCompleted(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	only := f.workout(plan, 0, "planned")
	f.completeWithHardFeedback(only)
	if status := f.planStatus(plan); status != "completed" {
		t.Fatalf("precondition: finishing the last workout completes the plan, got %s", status)
	}
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, only); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if status := f.planStatus(plan); status != "active" {
		t.Errorf("plan is %s, want active again", status)
	}
	if r := f.row(only); r.status != "planned" {
		t.Errorf("workout is %s, want planned", r.status)
	}
}

func (f *undoFixture) planStatus(planID string) string {
	var status string
	f.scan(&status, `SELECT status FROM training_plans WHERE id = $1`, planID)
	return status
}

func TestUndoRefusesWhatDoesNotBelongToTheCurrentPlan(t *testing.T) {
	f := newUndoFixture(t)

	oldPlan := f.plan("completed")
	oldWorkout := f.workout(oldPlan, -10, "completed")
	f.exec(`UPDATE training_plans SET created_at = now() - interval '2 days' WHERE id = $1`, oldPlan)
	f.plan("active")
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, oldWorkout); !errors.Is(err, planning.ErrWorkoutMissing) {
		t.Errorf("a workout of an old plan returned %v, want ErrWorkoutMissing", err)
	}

	other := newUndoFixture(t)
	otherPlan := other.plan("active")
	foreign := other.workout(otherPlan, -1, "completed")
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, foreign); !errors.Is(err, planning.ErrWorkoutMissing) {
		t.Errorf("another user's workout returned %v, want ErrWorkoutMissing", err)
	}
}

func TestUndoReopensASkippedWorkout(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	other := f.workout(plan, 1, "planned")
	skipped := f.workout(plan, -1, "skipped")
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, skipped); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if r := f.row(skipped); r.status != "planned" {
		t.Errorf("skipped workout is %s, want planned", r.status)
	}
	if r := f.row(other); r.status != "planned" {
		t.Errorf("other workout changed: %+v", r)
	}
}

func TestStartIsRefusedForAFutureWorkout(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	future := f.workout(plan, 2, "planned")
	today := f.workout(plan, 0, "planned")
	past := f.workout(plan, -2, "planned")
	date := time.Now().UTC().Format("2006-01-02")

	if err := f.store.StartWorkoutByUserID(f.ctx, f.userID, future, date); !errors.Is(err, planning.ErrWorkoutInFuture) {
		t.Errorf("a future workout returned %v, want ErrWorkoutInFuture", err)
	}
	if r := f.row(future); r.status != "planned" {
		t.Errorf("the refused workout changed to %s", r.status)
	}
	if n := f.count(`SELECT count(*) FROM workout_sessions WHERE workout_id = $1`, future); n != 0 {
		t.Errorf("a session was created for a future workout")
	}
	for name, id := range map[string]string{"hoje": today, "passado": past} {
		if err := f.store.StartWorkoutByUserID(f.ctx, f.userID, id, date); err != nil {
			t.Errorf("start %s: %v", name, err)
		}
	}
	// In the evening in Brazil the UTC date is already tomorrow; the client's date wins.
	evening := f.workout(plan, 1, "planned")
	tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
	yesterdayForClient := time.Now().UTC().Format("2006-01-02")
	if err := f.store.StartWorkoutByUserID(f.ctx, f.userID, evening, yesterdayForClient); !errors.Is(err, planning.ErrWorkoutInFuture) {
		t.Errorf("tomorrow's workout with a client date of today returned %v, want ErrWorkoutInFuture", err)
	}
	if err := f.store.StartWorkoutByUserID(f.ctx, f.userID, evening, tomorrow); err != nil {
		t.Errorf("tomorrow's workout when the client's date is tomorrow: %v", err)
	}
}
