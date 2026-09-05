package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
	"github.com/jackc/pgx/v5"
)

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
