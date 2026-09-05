package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/jackc/pgx/v5"
)

const planningTrainingHistoryQuery = `
	WITH window_sizes(window_days) AS (VALUES (7), (28), (42)),
	performed AS (
		SELECT window_sizes.window_days,
			COUNT(ws.id) AS performed_sessions,
			COALESCE(SUM(COALESCE(ws.duration_minutes, 0)), 0) AS performed_minutes,
			COUNT(ws.id) FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10) AS sessions_with_load,
			COUNT(ws.id) - COUNT(ws.id) FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10) AS sessions_without_load,
			COALESCE(SUM(ws.duration_minutes * ws.actual_rpe)
				FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10), 0)::double precision AS session_rpe_load,
			COUNT(ws.id) FILTER (WHERE f.id IS NOT NULL) AS feedback_records,
			COUNT(ws.id) FILTER (WHERE f.id IS NOT NULL AND f.fatigue_after BETWEEN 1 AND 5) AS sessions_with_complete_feedback,
			COUNT(ws.id) FILTER (WHERE f.pain_reported = true) AS pain_reported_sessions,
			COUNT(ws.id) FILTER (WHERE f.fatigue_after BETWEEN 4 AND 5) AS high_fatigue_sessions,
			COUNT(ws.id) FILTER (WHERE ws.actual_rpe BETWEEN 1 AND 10 AND source_workout.target_rpe IS NOT NULL
				AND ws.actual_rpe >= source_workout.target_rpe + 2) AS above_target_rpe_sessions
		FROM window_sizes
		LEFT JOIN workout_sessions ws
			ON ws.athlete_profile_id = $1
			AND ws.status = 'completed'
			AND ws.completed_at >= now() - make_interval(days => window_sizes.window_days)
			AND ws.completed_at <= now()
		LEFT JOIN workouts source_workout ON source_workout.id = ws.workout_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		GROUP BY window_sizes.window_days
	),
	recovery AS (
		SELECT window_sizes.window_days,
			COUNT(rd.id) AS recovery_checkins,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5) AS complete_recovery_checkins,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5
				AND (rd.sleep_minutes < 360 OR rd.sleep_quality <= 2 OR rd.stress_level >= 4 OR rd.fatigue_level >= 4)) AS checkins_with_protective_signal,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5
				AND (rd.fatigue_level = 5 OR
					(CASE WHEN rd.sleep_minutes < 360 OR rd.sleep_quality <= 2 THEN 1 ELSE 0 END +
					 CASE WHEN rd.stress_level >= 4 THEN 1 ELSE 0 END +
					 CASE WHEN rd.fatigue_level >= 4 THEN 1 ELSE 0 END) >= 2)) AS recovery_needed_checkins
		FROM window_sizes
		LEFT JOIN recovery_data rd
			ON rd.athlete_profile_id = $1
			AND rd.recorded_on >= CURRENT_DATE - (window_sizes.window_days - 1)
			AND rd.recorded_on <= CURRENT_DATE
		GROUP BY window_sizes.window_days
	),
	temporal AS (
		SELECT
			MAX(ws.completed_at) FILTER (WHERE ws.status = 'completed' AND ws.completed_at <= now()) AS latest_completed_at,
			MAX(ws.completed_at) FILTER (WHERE ws.status = 'completed' AND ws.completed_at <= now()
				AND ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10) AS latest_session_rpe_load_at,
			COUNT(ws.id) FILTER (WHERE ws.status = 'completed' AND ws.completed_at > now()) AS future_completed_sessions_excluded
		FROM workout_sessions ws
		WHERE ws.athlete_profile_id = $1
	),
	recovery_temporal AS (
		SELECT
			MAX(rd.recorded_on) FILTER (WHERE rd.recorded_on <= CURRENT_DATE) AS latest_recovery_recorded_on,
			COUNT(rd.id) FILTER (WHERE rd.recorded_on > CURRENT_DATE) AS future_recovery_checkins_excluded
		FROM recovery_data rd
		WHERE rd.athlete_profile_id = $1
	),
	adherence AS (
		SELECT window_sizes.window_days,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE OR w.status IN ('completed', 'skipped')) AS expected_sessions,
			COUNT(w.id) FILTER (WHERE w.status = 'completed') AS scheduled_completed_sessions,
			COUNT(w.id) FILTER (WHERE w.status = 'skipped') AS cancelled_sessions,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE AND w.status IN ('planned', 'adapted')) AS missed_sessions,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE AND w.status = 'in_progress') AS overdue_in_progress_sessions
		FROM window_sizes
		LEFT JOIN training_plans tp
			ON tp.athlete_profile_id = $1 AND tp.status IN ('active', 'completed')
		LEFT JOIN workouts w
			ON w.training_plan_id = tp.id
			AND w.scheduled_on >= CURRENT_DATE - (window_sizes.window_days - 1)
			AND w.scheduled_on <= CURRENT_DATE
		GROUP BY window_sizes.window_days
	)
	SELECT performed.window_days, adherence.expected_sessions, adherence.scheduled_completed_sessions,
		adherence.cancelled_sessions, adherence.missed_sessions, adherence.overdue_in_progress_sessions,
		performed.performed_sessions, performed.performed_minutes, performed.sessions_with_load,
		performed.sessions_without_load, performed.session_rpe_load, performed.feedback_records,
		performed.sessions_with_complete_feedback, performed.pain_reported_sessions,
		performed.high_fatigue_sessions, performed.above_target_rpe_sessions,
		recovery.recovery_checkins, recovery.complete_recovery_checkins,
		recovery.checkins_with_protective_signal, recovery.recovery_needed_checkins,
		temporal.latest_completed_at,
		CASE WHEN temporal.latest_completed_at IS NULL THEN NULL
			ELSE FLOOR(EXTRACT(EPOCH FROM (now() - temporal.latest_completed_at)) / 86400)::integer END,
		temporal.latest_session_rpe_load_at,
		CASE WHEN temporal.latest_session_rpe_load_at IS NULL THEN NULL
			ELSE FLOOR(EXTRACT(EPOCH FROM (now() - temporal.latest_session_rpe_load_at)) / 86400)::integer END,
		recovery_temporal.latest_recovery_recorded_on,
		CASE WHEN recovery_temporal.latest_recovery_recorded_on IS NULL THEN NULL
			ELSE CURRENT_DATE - recovery_temporal.latest_recovery_recorded_on END,
		temporal.future_completed_sessions_excluded, recovery_temporal.future_recovery_checkins_excluded
	FROM performed
	JOIN adherence USING (window_days)
	JOIN recovery USING (window_days)
	CROSS JOIN temporal
	CROSS JOIN recovery_temporal
	ORDER BY performed.window_days`

const planningTrainingHistoryPeriodsQuery = `
	WITH period_sizes(period_index, period_key, period_days, start_days_ago, end_days_ago) AS (
		VALUES
			(0, 'last_7d', 7, 0, 7),
			(1, 'days_8_14', 7, 7, 14),
			(2, 'days_15_21', 7, 14, 21),
			(3, 'days_22_28', 7, 21, 28),
			(4, 'days_29_35', 7, 28, 35),
			(5, 'days_36_42', 7, 35, 42)
	),
	performed AS (
		SELECT period_sizes.period_index,
			period_sizes.period_key,
			period_sizes.period_days,
			COUNT(ws.id) AS performed_sessions,
			COALESCE(SUM(COALESCE(ws.duration_minutes, 0)), 0) AS performed_minutes,
			COUNT(ws.id) FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10) AS sessions_with_load,
			COUNT(ws.id) - COUNT(ws.id) FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10) AS sessions_without_load,
			COALESCE(SUM(ws.duration_minutes * ws.actual_rpe)
				FILTER (WHERE ws.duration_minutes > 0 AND ws.actual_rpe BETWEEN 1 AND 10), 0)::double precision AS session_rpe_load,
			COUNT(ws.id) FILTER (WHERE f.id IS NOT NULL) AS feedback_records,
			COUNT(ws.id) FILTER (WHERE f.id IS NOT NULL AND f.fatigue_after BETWEEN 1 AND 5) AS sessions_with_complete_feedback,
			COUNT(ws.id) FILTER (WHERE f.pain_reported = true) AS pain_reported_sessions,
			COUNT(ws.id) FILTER (WHERE f.fatigue_after BETWEEN 4 AND 5) AS high_fatigue_sessions,
			COUNT(ws.id) FILTER (WHERE ws.actual_rpe BETWEEN 1 AND 10 AND source_workout.target_rpe IS NOT NULL
				AND ws.actual_rpe >= source_workout.target_rpe + 2) AS above_target_rpe_sessions
		FROM period_sizes
		LEFT JOIN workout_sessions ws
			ON ws.athlete_profile_id = $1
			AND ws.status = 'completed'
			AND ws.completed_at >= now() - make_interval(days => period_sizes.end_days_ago)
			AND ws.completed_at < now() - make_interval(days => period_sizes.start_days_ago)
		LEFT JOIN workouts source_workout ON source_workout.id = ws.workout_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		GROUP BY period_sizes.period_index, period_sizes.period_key, period_sizes.period_days
	),
	recovery AS (
		SELECT period_sizes.period_index,
			COUNT(rd.id) AS recovery_checkins,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5) AS complete_recovery_checkins,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5
				AND (rd.sleep_minutes < 360 OR rd.sleep_quality <= 2 OR rd.stress_level >= 4 OR rd.fatigue_level >= 4)) AS checkins_with_protective_signal,
			COUNT(rd.id) FILTER (WHERE rd.sleep_minutes >= 0 AND rd.sleep_quality BETWEEN 1 AND 5
				AND rd.stress_level BETWEEN 1 AND 5 AND rd.fatigue_level BETWEEN 1 AND 5
				AND (rd.fatigue_level = 5 OR
					(CASE WHEN rd.sleep_minutes < 360 OR rd.sleep_quality <= 2 THEN 1 ELSE 0 END +
					 CASE WHEN rd.stress_level >= 4 THEN 1 ELSE 0 END +
					 CASE WHEN rd.fatigue_level >= 4 THEN 1 ELSE 0 END) >= 2)) AS recovery_needed_checkins
		FROM period_sizes
		LEFT JOIN recovery_data rd
			ON rd.athlete_profile_id = $1
			AND rd.recorded_on >= CURRENT_DATE - (period_sizes.end_days_ago - 1)
			AND rd.recorded_on < CURRENT_DATE - (period_sizes.start_days_ago - 1)
		GROUP BY period_sizes.period_index
	),
	adherence AS (
		SELECT period_sizes.period_index,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE OR w.status IN ('completed', 'skipped')) AS expected_sessions,
			COUNT(w.id) FILTER (WHERE w.status = 'completed') AS scheduled_completed_sessions,
			COUNT(w.id) FILTER (WHERE w.status = 'skipped') AS cancelled_sessions,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE AND w.status IN ('planned', 'adapted')) AS missed_sessions,
			COUNT(w.id) FILTER (WHERE w.scheduled_on < CURRENT_DATE AND w.status = 'in_progress') AS overdue_in_progress_sessions
		FROM period_sizes
		LEFT JOIN training_plans tp
			ON tp.athlete_profile_id = $1 AND tp.status IN ('active', 'completed')
		LEFT JOIN workouts w
			ON w.training_plan_id = tp.id
			AND w.scheduled_on >= CURRENT_DATE - (period_sizes.end_days_ago - 1)
			AND w.scheduled_on < CURRENT_DATE - (period_sizes.start_days_ago - 1)
		GROUP BY period_sizes.period_index
	)
	SELECT performed.period_index, performed.period_key, performed.period_days,
		adherence.expected_sessions, adherence.scheduled_completed_sessions,
		adherence.cancelled_sessions, adherence.missed_sessions, adherence.overdue_in_progress_sessions,
		performed.performed_sessions, performed.performed_minutes, performed.sessions_with_load,
		performed.sessions_without_load, performed.session_rpe_load, performed.feedback_records,
		performed.sessions_with_complete_feedback, performed.pain_reported_sessions,
		performed.high_fatigue_sessions, performed.above_target_rpe_sessions,
		recovery.recovery_checkins, recovery.complete_recovery_checkins,
		recovery.checkins_with_protective_signal, recovery.recovery_needed_checkins
	FROM performed
	JOIN adherence USING (period_index)
	JOIN recovery USING (period_index)
	ORDER BY performed.period_index`

func (s *Store) PlanningContextByUserID(ctx context.Context, userID string) (planning.Context, error) {
	var input planning.Context
	var cyclingContext []byte
	err := s.pool.QueryRow(ctx, `
		SELECT ap.id::text, ap.experience_level, ap.cycling_context,
			COALESCE((SELECT ca.eligible_for_progression FROM cycling_assessments ca WHERE ca.athlete_profile_id = ap.id ORDER BY ca.completed_at DESC LIMIT 1), false)
		FROM athlete_profiles ap WHERE ap.user_id = $1`, userID,
	).Scan(&input.ProfileID, &input.ExperienceLevel, &cyclingContext, &input.BaselineEligible)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.Context{}, planning.ErrIncompleteOnboarding
	}
	if err != nil {
		return planning.Context{}, err
	}
	if err := json.Unmarshal(cyclingContext, &input.Cycling); err != nil {
		return planning.Context{}, err
	}
	input.Observed.WindowDays = 28
	input.Observed.DataCoverage = &planning.ObservedDataCoverage{}
	var completedSessions int64
	if err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE ws.status = 'completed'),
			COALESCE(SUM(ws.duration_minutes) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.actual_rpe) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(AVG(f.fatigue_after) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(BOOL_OR(f.pain_reported) FILTER (WHERE ws.status = 'completed'), false),
			COUNT(*) FILTER (WHERE ws.status = 'completed' AND ws.duration_minutes > 0),
			COUNT(*) FILTER (WHERE ws.status = 'completed' AND ws.actual_rpe BETWEEN 1 AND 10),
			COUNT(*) FILTER (WHERE ws.status = 'completed' AND f.fatigue_after BETWEEN 1 AND 5 AND f.pain_reported IS NOT NULL),
			COUNT(*) FILTER (WHERE ws.status = 'completed' AND ws.duration_minutes > 0
				AND ws.actual_rpe BETWEEN 1 AND 10 AND f.fatigue_after BETWEEN 1 AND 5 AND f.pain_reported IS NOT NULL)
		FROM workout_sessions ws
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ws.athlete_profile_id = $1
		  AND ws.completed_at >= now() - interval '28 days'`, input.ProfileID,
	).Scan(&completedSessions, &input.Observed.CompletedMinutes, &input.Observed.AverageRPE,
		&input.Observed.AverageFatigue, &input.Observed.PainReported,
		&input.Observed.DataCoverage.SessionsWithDuration, &input.Observed.DataCoverage.SessionsWithRPE,
		&input.Observed.DataCoverage.SessionsWithFeedback, &input.Observed.DataCoverage.CompleteSessions); err != nil {
		return planning.Context{}, err
	}
	input.Observed.CompletedSessions = int(completedSessions)
	var recoveryCheckins int64
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(AVG(fatigue_level), 0)::double precision,
			COUNT(*) FILTER (WHERE fatigue_level BETWEEN 1 AND 5)
		FROM recovery_data
		WHERE athlete_profile_id = $1
		  AND recorded_on >= CURRENT_DATE - 27`, input.ProfileID,
	).Scan(&recoveryCheckins, &input.Observed.AverageRecoveryFatigue, &input.Observed.DataCoverage.RecoveryWithFatigue); err != nil {
		return planning.Context{}, err
	}
	input.Observed.RecoveryCheckins = int(recoveryCheckins)
	input.TrainingHistory, err = s.trainingHistoryByProfileID(ctx, input.ProfileID)
	if err != nil {
		return planning.Context{}, err
	}
	input.TrainingHistoryPeriods, err = s.trainingHistoryPeriodsByProfileID(ctx, input.ProfileID)
	if err != nil {
		return planning.Context{}, err
	}
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM training_plans WHERE athlete_profile_id = $1`, input.ProfileID).Scan(&input.RotationIndex); err != nil {
		return planning.Context{}, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT goal_type FROM goals WHERE athlete_profile_id = $1 AND priority = 1`, input.ProfileID,
	).Scan(&input.PrimaryGoal)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.Context{}, planning.ErrIncompleteOnboarding
	}
	if err != nil {
		return planning.Context{}, err
	}

	limitationRows, err := s.pool.Query(ctx, `
		SELECT kind, professional_clearance_recommended FROM injuries_or_limitations
		WHERE athlete_profile_id = $1 AND is_active = true`, input.ProfileID)
	if err != nil {
		return planning.Context{}, err
	}
	for limitationRows.Next() {
		var item planning.LimitationContext
		if err := limitationRows.Scan(&item.Kind, &item.ProfessionalClearanceRecommended); err != nil {
			limitationRows.Close()
			return planning.Context{}, err
		}
		input.Limitations = append(input.Limitations, item)
	}
	if err := limitationRows.Err(); err != nil {
		limitationRows.Close()
		return planning.Context{}, err
	}
	limitationRows.Close()

	availabilityRows, err := s.pool.Query(ctx, `
		SELECT weekday, available_minutes, location FROM availability
		WHERE athlete_profile_id = $1 AND available_minutes > 0 ORDER BY weekday`, input.ProfileID)
	if err != nil {
		return planning.Context{}, err
	}
	defer availabilityRows.Close()
	for availabilityRows.Next() {
		var item planning.AvailabilitySlot
		if err := availabilityRows.Scan(&item.Weekday, &item.AvailableMinutes, &item.Location); err != nil {
			return planning.Context{}, err
		}
		input.Availability = append(input.Availability, item)
	}
	if err := availabilityRows.Err(); err != nil {
		return planning.Context{}, err
	}
	if len(input.Availability) == 0 {
		return planning.Context{}, planning.ErrIncompleteOnboarding
	}
	return input, nil
}

func (s *Store) trainingHistoryByProfileID(ctx context.Context, profileID string) ([]planning.TrainingHistoryWindow, error) {
	rows, err := s.pool.Query(ctx, planningTrainingHistoryQuery, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	history := make([]planning.TrainingHistoryWindow, 0, 3)
	for rows.Next() {
		var window planning.TrainingHistoryWindow
		if err := rows.Scan(
			&window.WindowDays, &window.ExpectedSessions, &window.ScheduledCompletedSessions,
			&window.CancelledSessions, &window.MissedSessions, &window.OverdueInProgressSessions,
			&window.PerformedSessions, &window.PerformedMinutes, &window.SessionsWithSessionRPELoad,
			&window.SessionsWithoutSessionRPELoad, &window.SessionRPELoad, &window.FeedbackRecords,
			&window.SessionsWithCompleteFeedback, &window.PainReportedSessions,
			&window.HighFatigueSessions, &window.AboveTargetRPESessions,
			&window.RecoveryCheckins, &window.CompleteRecoveryCheckins,
			&window.CheckinsWithProtectiveSignal, &window.RecoveryNeededCheckins,
			&window.LatestCompletedAt, &window.DaysSinceLatestCompleted,
			&window.LatestSessionRPELoadAt, &window.DaysSinceLatestSessionRPELoad,
			&window.LatestRecoveryRecordedOn, &window.DaysSinceLatestRecoveryCheckin,
			&window.FutureCompletedSessionsExcluded, &window.FutureRecoveryCheckinsExcluded,
		); err != nil {
			return nil, err
		}
		history = append(history, window)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return history, nil
}

func (s *Store) trainingHistoryPeriodsByProfileID(ctx context.Context, profileID string) ([]planning.TrainingHistoryPeriod, error) {
	rows, err := s.pool.Query(ctx, planningTrainingHistoryPeriodsQuery, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	periods := make([]planning.TrainingHistoryPeriod, 0, 6)
	for rows.Next() {
		var period planning.TrainingHistoryPeriod
		if err := rows.Scan(
			&period.PeriodIndex, &period.PeriodKey, &period.PeriodDays,
			&period.ExpectedSessions, &period.ScheduledCompletedSessions, &period.CancelledSessions,
			&period.MissedSessions, &period.OverdueInProgressSessions,
			&period.PerformedSessions, &period.PerformedMinutes, &period.SessionsWithSessionRPELoad,
			&period.SessionsWithoutSessionRPELoad, &period.SessionRPELoad, &period.FeedbackRecords,
			&period.SessionsWithCompleteFeedback, &period.PainReportedSessions,
			&period.HighFatigueSessions, &period.AboveTargetRPESessions,
			&period.RecoveryCheckins, &period.CompleteRecoveryCheckins,
			&period.CheckinsWithProtectiveSignal, &period.RecoveryNeededCheckins,
		); err != nil {
			return nil, err
		}
		periods = append(periods, period)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return periods, nil
}

func (s *Store) SaveDraftPlan(ctx context.Context, profileID string, plan planning.Plan) (planning.Plan, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return planning.Plan{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM training_plans WHERE athlete_profile_id = $1 AND status = 'draft'`, profileID); err != nil {
		return planning.Plan{}, err
	}
	snapshot, err := json.Marshal(plan.PrescriptionSnapshot)
	if err != nil {
		return planning.Plan{}, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status, prescription_snapshot)
		VALUES ($1, $2::date, $3::date, 'draft', $4::jsonb) RETURNING id::text`,
		profileID, plan.StartsOn, plan.EndsOn, snapshot,
	).Scan(&plan.ID)
	if err != nil {
		return planning.Plan{}, err
	}
	for index := range plan.Workouts {
		structure, err := json.Marshal(plan.Workouts[index].Structure)
		if err != nil {
			return planning.Plan{}, err
		}
		explanation, err := json.Marshal(plan.Workouts[index].Explanation)
		if err != nil {
			return planning.Plan{}, err
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
			VALUES ($1, $2::date, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9) RETURNING id::text`,
			plan.ID, plan.Workouts[index].ScheduledOn, plan.Workouts[index].Name, plan.Workouts[index].Objective,
			plan.Workouts[index].DurationMinutes, plan.Workouts[index].TargetRPE, structure, explanation, plan.Workouts[index].Status,
		).Scan(&plan.Workouts[index].ID)
		if err != nil {
			return planning.Plan{}, err
		}
	}
	evidenceRows, err := tx.Query(ctx, `SELECT source_key, title, authors, published_year, url, training_focus, evidence_level, summary FROM scientific_sources ORDER BY source_key`)
	if err != nil {
		return planning.Plan{}, err
	}
	defer evidenceRows.Close()
	for evidenceRows.Next() {
		var source planning.ScientificSource
		if err := evidenceRows.Scan(&source.SourceKey, &source.Title, &source.Authors, &source.PublishedYear, &source.URL, &source.TrainingFocus, &source.EvidenceLevel, &source.Summary); err != nil {
			return planning.Plan{}, err
		}
		plan.Evidence = append(plan.Evidence, source)
	}
	if err := evidenceRows.Err(); err != nil {
		return planning.Plan{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return planning.Plan{}, err
	}
	return plan, nil
}

func (s *Store) CurrentPlanByUserID(ctx context.Context, userID string) (planning.Plan, error) {
	var plan planning.Plan
	var snapshot, cyclingContext []byte
	err := s.pool.QueryRow(ctx, `
		SELECT tp.id::text, tp.starts_on::text, tp.ends_on::text, tp.status, tp.prescription_snapshot, ap.cycling_context
		FROM training_plans tp JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
		WHERE ap.user_id = $1 AND tp.status IN ('active', 'draft', 'completed')
		ORDER BY CASE tp.status WHEN 'draft' THEN 0 WHEN 'active' THEN 1 ELSE 2 END, tp.created_at DESC LIMIT 1`, userID,
	).Scan(&plan.ID, &plan.StartsOn, &plan.EndsOn, &plan.Status, &snapshot, &cyclingContext)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.Plan{}, planning.ErrPlanMissing
	}
	if err != nil {
		return planning.Plan{}, err
	}
	if err := json.Unmarshal(snapshot, &plan.PrescriptionSnapshot); err != nil {
		return planning.Plan{}, err
	}
	var currentCyclingContext map[string]any
	if err := json.Unmarshal(cyclingContext, &currentCyclingContext); err != nil {
		return planning.Plan{}, err
	}
	plan.PrescriptionSnapshot["cycling_context"] = currentCyclingContext
	plan.Workouts = []planning.Workout{}
	rows, err := s.pool.Query(ctx, `
		SELECT w.id::text, w.scheduled_on::text, w.name, w.objective, w.duration_minutes,
			w.target_rpe::double precision, w.structure, w.explanation, w.status,
			ws.id::text, ws.status, ws.started_at, ws.completed_at, ws.cancelled_at,
			ws.duration_minutes, ws.actual_rpe::double precision, ws.distance_km::double precision, ws.elevation_gain_m,
			ws.average_power_watts, ws.average_heart_rate,
			f.difficulty, f.pain_reported, f.fatigue_after, f.notes
		FROM workouts w
		LEFT JOIN LATERAL (
			SELECT latest.*
			FROM workout_sessions latest
			WHERE latest.workout_id = w.id
			ORDER BY latest.created_at DESC
			LIMIT 1
		) ws ON true
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE w.training_plan_id = $1
		ORDER BY w.scheduled_on, w.created_at`, plan.ID)
	if err != nil {
		return planning.Plan{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var workout planning.Workout
		var structure, explanation []byte
		var sessionID, sessionStatus, difficulty, notes *string
		var startedAt, completedAt, cancelledAt *time.Time
		var durationMinutes, fatigueAfter, elevationGainM, averagePowerW, averageHeartRate *int
		var actualRPE, distanceKM *float64
		var painReported *bool
		if err := rows.Scan(
			&workout.ID, &workout.ScheduledOn, &workout.Name, &workout.Objective,
			&workout.DurationMinutes, &workout.TargetRPE, &structure, &explanation, &workout.Status,
			&sessionID, &sessionStatus, &startedAt, &completedAt, &cancelledAt,
			&durationMinutes, &actualRPE, &distanceKM, &elevationGainM, &averagePowerW, &averageHeartRate,
			&difficulty, &painReported, &fatigueAfter, &notes,
		); err != nil {
			return planning.Plan{}, err
		}
		if err := json.Unmarshal(structure, &workout.Structure); err != nil {
			return planning.Plan{}, err
		}
		if err := json.Unmarshal(explanation, &workout.Explanation); err != nil {
			return planning.Plan{}, err
		}
		if sessionID != nil && sessionStatus != nil {
			workout.Session = &planning.WorkoutSession{
				ID: *sessionID, Status: *sessionStatus, StartedAt: startedAt,
				CompletedAt: completedAt, CancelledAt: cancelledAt,
				DurationMinutes: durationMinutes, ActualRPE: actualRPE, DistanceKM: distanceKM,
				ElevationGainM: elevationGainM, AveragePowerW: averagePowerW, AverageHeartRate: averageHeartRate,
			}
			if difficulty != nil && painReported != nil && fatigueAfter != nil {
				workout.Session.Feedback = &planning.Feedback{
					Difficulty: *difficulty, PainReported: *painReported,
					FatigueAfter: *fatigueAfter,
				}
				if notes != nil {
					workout.Session.Feedback.Notes = *notes
				}
			}
		}
		plan.Workouts = append(plan.Workouts, workout)
	}
	if err := rows.Err(); err != nil {
		return planning.Plan{}, err
	}
	evidenceRows, err := s.pool.Query(ctx, `SELECT source_key, title, authors, published_year, url, training_focus, evidence_level, summary FROM scientific_sources ORDER BY source_key`)
	if err != nil {
		return planning.Plan{}, err
	}
	defer evidenceRows.Close()
	for evidenceRows.Next() {
		var source planning.ScientificSource
		if err := evidenceRows.Scan(&source.SourceKey, &source.Title, &source.Authors, &source.PublishedYear, &source.URL, &source.TrainingFocus, &source.EvidenceLevel, &source.Summary); err != nil {
			return planning.Plan{}, err
		}
		plan.Evidence = append(plan.Evidence, source)
	}
	if err := evidenceRows.Err(); err != nil {
		return planning.Plan{}, err
	}
	return plan, nil
}

func (s *Store) ActivatePlanByUserID(ctx context.Context, userID, planID string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var profileID string
	err = tx.QueryRow(ctx, `
		SELECT ap.id::text
		FROM athlete_profiles ap
		JOIN training_plans tp ON tp.athlete_profile_id = ap.id
		WHERE ap.user_id = $1 AND tp.id = $2 AND tp.status = 'draft'
		FOR UPDATE OF ap, tp`, userID, planID,
	).Scan(&profileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return planning.ErrPlanMissing
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE training_plans SET status = 'cancelled'
		WHERE athlete_profile_id = $1 AND status = 'active'`, profileID); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `
		UPDATE training_plans SET status = 'active'
		WHERE id = $1 AND athlete_profile_id = $2 AND status = 'draft'`, planID, profileID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return planning.ErrPlanMissing
	}
	return tx.Commit(ctx)
}
