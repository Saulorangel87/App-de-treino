package repository

import (
	"context"

	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
)

func (s *Store) EvolutionSummaryByUserID(ctx context.Context, userID string) (evolution.Summary, error) {
	result := evolution.Summary{Weeks: make([]evolution.Week, 0, 8), Recovery: make([]evolution.RecoveryPoint, 0, 14)}
	var completed, cancelled int64
	if err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE ws.status = 'completed'),
			COUNT(*) FILTER (WHERE ws.status = 'cancelled'),
			COALESCE(SUM(ws.duration_minutes) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.actual_rpe) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(AVG(f.fatigue_after) FILTER (WHERE ws.status = 'completed'), 0)::double precision
		FROM workout_sessions ws
		JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ap.user_id = $1`, userID,
	).Scan(&completed, &cancelled, &result.TotalMinutes, &result.AverageRPE, &result.AverageFatigue); err != nil {
		return evolution.Summary{}, err
	}
	result.CompletedSessions, result.CancelledSessions = int(completed), int(cancelled)
	if completed+cancelled > 0 {
		result.CompletionRate = float64(completed) / float64(completed+cancelled) * 100
	}

	weeks, err := s.pool.Query(ctx, `
		WITH weeks AS (
			SELECT value::date AS week_start
			FROM generate_series(
				date_trunc('week', CURRENT_DATE)::date - 49,
				date_trunc('week', CURRENT_DATE)::date,
				interval '7 days'
			) AS value
		)
		SELECT w.week_start::text,
			COUNT(ws.id) FILTER (WHERE ws.status = 'completed'),
			COUNT(ws.id) FILTER (WHERE ws.status = 'cancelled'),
			COALESCE(SUM(ws.duration_minutes) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.actual_rpe) FILTER (WHERE ws.status = 'completed'), 0)::double precision
		FROM weeks w
		LEFT JOIN athlete_profiles ap ON ap.user_id = $1
		LEFT JOIN workout_sessions ws ON ws.athlete_profile_id = ap.id
			AND COALESCE(ws.completed_at, ws.cancelled_at)::date >= w.week_start
			AND COALESCE(ws.completed_at, ws.cancelled_at)::date < w.week_start + 7
		GROUP BY w.week_start
		ORDER BY w.week_start`, userID)
	if err != nil {
		return evolution.Summary{}, err
	}
	defer weeks.Close()
	for weeks.Next() {
		var week evolution.Week
		var weekCompleted, weekCancelled int64
		if err := weeks.Scan(&week.WeekStart, &weekCompleted, &weekCancelled, &week.TotalMinutes, &week.AverageRPE); err != nil {
			return evolution.Summary{}, err
		}
		week.CompletedSessions, week.CancelledSessions = int(weekCompleted), int(weekCancelled)
		result.Weeks = append(result.Weeks, week)
	}
	if err := weeks.Err(); err != nil {
		return evolution.Summary{}, err
	}

	recoveryRows, err := s.pool.Query(ctx, `
		SELECT rd.recorded_on::text, COALESCE(rd.sleep_minutes, 0), COALESCE(rd.sleep_quality, 3),
			COALESCE(rd.stress_level, 3), COALESCE(rd.fatigue_level, 3),
			CASE
				WHEN COALESCE(rd.fatigue_level, 3) = 5 OR
					(CASE WHEN COALESCE(rd.sleep_minutes, 0) < 360 OR COALESCE(rd.sleep_quality, 3) <= 2 THEN 1 ELSE 0 END +
					 CASE WHEN COALESCE(rd.stress_level, 3) >= 4 THEN 1 ELSE 0 END +
					 CASE WHEN COALESCE(rd.fatigue_level, 3) >= 4 THEN 1 ELSE 0 END) >= 2
				THEN 'recovery_needed'
				WHEN COALESCE(rd.sleep_minutes, 0) < 360 OR COALESCE(rd.sleep_quality, 3) <= 2 OR
					COALESCE(rd.stress_level, 3) >= 4 OR COALESCE(rd.fatigue_level, 3) >= 4
				THEN 'caution'
				ELSE 'ready'
			END
		FROM recovery_data rd
		JOIN athlete_profiles ap ON ap.id = rd.athlete_profile_id
		WHERE ap.user_id = $1
		ORDER BY rd.recorded_on DESC
		LIMIT 14`, userID)
	if err != nil {
		return evolution.Summary{}, err
	}
	defer recoveryRows.Close()
	for recoveryRows.Next() {
		var point evolution.RecoveryPoint
		if err := recoveryRows.Scan(&point.RecordedOn, &point.SleepMinutes, &point.SleepQuality, &point.StressLevel, &point.FatigueLevel, &point.Readiness); err != nil {
			return evolution.Summary{}, err
		}
		result.Recovery = append(result.Recovery, point)
	}
	if err := recoveryRows.Err(); err != nil {
		return evolution.Summary{}, err
	}
	return result, nil
}
