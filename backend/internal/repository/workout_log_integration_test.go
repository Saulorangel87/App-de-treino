package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

// Runs only when CADENCIA_TEST_DATABASE_URL points at a migrated database (000036).

func easyFeedback() planning.CompletionInput {
	return planning.CompletionInput{CompletionStatus: "complete", ActualRPE: 4, Difficulty: "easy", FatigueAfter: 2}
}

func hardFeedback() planning.CompletionInput {
	return planning.CompletionInput{CompletionStatus: "complete", ActualRPE: 8.5, Difficulty: "very_hard", FatigueAfter: 5}
}

func today() string { return time.Now().UTC().Format("2006-01-02") }

type sessionRow struct {
	status, source string
	duration       int
	minutes        float64 // completed_at - started_at, in minutes
	completedOn    string
	feedbacks      int
}

func (f *undoFixture) session(workoutID string) sessionRow {
	var r sessionRow
	f.scan(&r.status, `SELECT status FROM workout_sessions WHERE workout_id = $1`, workoutID)
	f.scan(&r.source, `SELECT duration_source FROM workout_sessions WHERE workout_id = $1`, workoutID)
	f.scan(&r.duration, `SELECT duration_minutes FROM workout_sessions WHERE workout_id = $1`, workoutID)
	f.scan(&r.minutes, `SELECT EXTRACT(EPOCH FROM (completed_at - started_at)) / 60 FROM workout_sessions WHERE workout_id = $1`, workoutID)
	f.scan(&r.completedOn, `SELECT (completed_at AT TIME ZONE 'UTC')::date::text FROM workout_sessions WHERE workout_id = $1`, workoutID)
	f.scan(&r.feedbacks, `SELECT count(*) FROM feedback f JOIN workout_sessions ws ON ws.id = f.workout_session_id WHERE ws.workout_id = $1`, workoutID)
	return r
}

func TestLogRecordsAReportedSessionThroughTheSamePathAsTheStopwatch(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	done := f.workout(plan, 0, "planned")
	next1 := f.workout(plan, 1, "planned")
	next2 := f.workout(plan, 2, "planned")

	err := f.store.LogWorkoutByUserID(f.ctx, f.userID, done, planning.LogWorkoutInput{
		CompletionInput: hardFeedback(), DurationMinutes: 75, PerformedOn: today(),
	}, today())
	if err != nil {
		t.Fatalf("log: %v", err)
	}

	if r := f.row(done); r.status != "completed" {
		t.Errorf("workout is %q, want completed", r.status)
	}
	session := f.session(done)
	if session.status != "completed" || session.source != "reported" || session.duration != 75 || session.feedbacks != 1 {
		t.Errorf("unexpected session: %+v", session)
	}
	if session.minutes < 74.9 || session.minutes > 75.1 {
		t.Errorf("started_at must be completed_at minus the reported duration, got %.2f minutes", session.minutes)
	}
	var integrity string
	var eligible bool
	f.scan(&integrity, `SELECT explanation #>> '{data_integrity,status}' FROM workouts WHERE id = $1`, done)
	f.scan(&eligible, `SELECT (explanation #>> '{data_integrity,eligible_for_history}')::boolean FROM workouts WHERE id = $1`, done)
	if integrity != "valid" || !eligible {
		t.Errorf("a reported duration above zero must be valid and eligible for history, got %s eligible=%t", integrity, eligible)
	}
	for _, id := range []string{next1, next2} {
		if r := f.row(id); r.status != "adapted" || r.duration >= 60 || !r.adapted {
			t.Errorf("the hard session must reduce the next workouts like the stopwatch does: %+v", r)
		}
	}

	// The record can be undone like any other, restoring what it adapted.
	if err := f.store.UndoWorkoutByUserID(f.ctx, f.userID, done); err != nil {
		t.Fatalf("undo: %v", err)
	}
	if r := f.row(next1); r.status != "planned" || r.duration != 60 || r.adapted {
		t.Errorf("undo did not restore the adapted workout: %+v", r)
	}
}

func TestLogPlacesAnEarlierDayAtMiddayInBrasilia(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	id := f.workout(plan, -5, "planned")
	day := time.Now().UTC().AddDate(0, 0, -3).Format("2006-01-02")

	if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, id, planning.LogWorkoutInput{
		CompletionInput: easyFeedback(), DurationMinutes: 60, PerformedOn: day,
	}, today()); err != nil {
		t.Fatalf("log: %v", err)
	}
	if session := f.session(id); session.completedOn != day {
		t.Errorf("session dated %s, want %s", session.completedOn, day)
	}
}

func TestLogRefusesWhatIsNotAPlannedWorkoutOfTodayOrBefore(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	future := f.workout(plan, 3, "planned")
	scheduledToday := f.workout(plan, 0, "planned")
	inProgress := f.workout(plan, -1, "in_progress")
	completed := f.workout(plan, -2, "completed")
	input := func(performedOn string) planning.LogWorkoutInput {
		return planning.LogWorkoutInput{CompletionInput: easyFeedback(), DurationMinutes: 60, PerformedOn: performedOn}
	}

	if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, future, input(today()), today()); !errors.Is(err, planning.ErrWorkoutInFuture) {
		t.Errorf("a future workout returned %v, want ErrWorkoutInFuture", err)
	}
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, scheduledToday, input(yesterday), today()); !errors.Is(err, planning.ErrInvalidLog) {
		t.Errorf("a ride dated before the workout returned %v, want ErrInvalidLog", err)
	}
	for name, id := range map[string]string{"em andamento": inProgress, "já concluído": completed} {
		if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, id, input(today()), today()); !errors.Is(err, planning.ErrInvalidTransition) {
			t.Errorf("%s returned %v, want ErrInvalidTransition", name, err)
		}
	}
	other := newUndoFixture(t)
	foreign := other.workout(other.plan("active"), -1, "planned")
	if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, foreign, input(today()), today()); !errors.Is(err, planning.ErrWorkoutMissing) {
		t.Errorf("another user's workout returned %v, want ErrWorkoutMissing", err)
	}
	for _, id := range []string{future, scheduledToday, inProgress} {
		if n := f.count(`SELECT count(*) FROM workout_sessions WHERE workout_id = $1`, id); n != 0 {
			t.Errorf("a refused log left %d session(s) behind", n)
		}
	}
}

func TestProgressionNeedsMostOfThePlannedTimeWhetherReportedOrMeasured(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	short := f.workout(plan, -3, "planned")
	long := f.workout(plan, -2, "planned")
	next1 := f.workout(plan, 1, "planned")
	next2 := f.workout(plan, 2, "planned")
	f.exec(`UPDATE workouts SET duration_minutes = 90, target_rpe = 6 WHERE id IN ($1, $2)`, short, long)

	log := func(id string, minutes int) {
		t.Helper()
		if err := f.store.LogWorkoutByUserID(f.ctx, f.userID, id, planning.LogWorkoutInput{
			CompletionInput: easyFeedback(), DurationMinutes: minutes, PerformedOn: today(),
		}, today()); err != nil {
			t.Fatalf("log %d min: %v", minutes, err)
		}
	}
	log(short, 30)
	if r := f.row(next1); r.adapted || r.duration != 60 {
		t.Errorf("30 of 90 minutes must not progress the next workout: %+v", r)
	}
	log(long, 75)
	if r := f.row(next1); !r.adapted || r.duration != 63 {
		t.Errorf("75 of 90 minutes (83%%) must progress the next workout to 63 min: %+v", r)
	}
	if r := f.row(next2); r.adapted {
		t.Errorf("only the next planned workout is progressed: %+v", r)
	}
}

func TestTheStopwatchFlowStillRecordsATimerSession(t *testing.T) {
	f := newUndoFixture(t)
	plan := f.plan("active")
	id := f.workout(plan, 0, "planned")

	if err := f.store.StartWorkoutByUserID(f.ctx, f.userID, id, today()); err != nil {
		t.Fatalf("start: %v", err)
	}
	f.exec(`UPDATE workout_sessions SET started_at = now() - interval '45 minutes' WHERE workout_id = $1`, id)
	if err := f.store.CompleteWorkoutByUserID(f.ctx, f.userID, id, easyFeedback()); err != nil {
		t.Fatalf("complete: %v", err)
	}
	session := f.session(id)
	if session.source != "timer" || session.duration < 44 || session.duration > 46 || session.feedbacks != 1 {
		t.Errorf("unexpected stopwatch session: %+v", session)
	}
	if r := f.row(id); r.status != "completed" {
		t.Errorf("workout is %q, want completed", r.status)
	}
}
