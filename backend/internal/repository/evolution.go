package repository

import (
	"context"

	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
)

func (s *Store) EvolutionSummaryByUserID(ctx context.Context, userID string) (evolution.Summary, error) {
	result := evolution.Summary{Weeks: make([]evolution.Week, 0, 8), RecentSessions: make([]evolution.SessionComparison, 0, 12), Recovery: make([]evolution.RecoveryPoint, 0, 14)}
	var completed, cancelled int64
	if err := s.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE ws.status = 'completed'),
			COUNT(*) FILTER (WHERE ws.status = 'cancelled'),
			COALESCE(SUM(ws.duration_minutes) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.actual_rpe) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(AVG(f.fatigue_after) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(SUM(ws.distance_km) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(SUM(ws.elevation_gain_m) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.average_power_watts) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(AVG(ws.average_heart_rate) FILTER (WHERE ws.status = 'completed'), 0)::double precision
		FROM workout_sessions ws
		JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ap.user_id = $1`, userID,
	).Scan(&completed, &cancelled, &result.TotalMinutes, &result.AverageRPE, &result.AverageFatigue,
		&result.TotalDistanceKM, &result.TotalElevationM, &result.AveragePowerW, &result.AverageHeartRate); err != nil {
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
			COALESCE(AVG(ws.actual_rpe) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(SUM(ws.distance_km) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(SUM(ws.elevation_gain_m) FILTER (WHERE ws.status = 'completed'), 0),
			COALESCE(AVG(ws.average_power_watts) FILTER (WHERE ws.status = 'completed'), 0)::double precision,
			COALESCE(AVG(ws.average_heart_rate) FILTER (WHERE ws.status = 'completed'), 0)::double precision
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
		if err := weeks.Scan(&week.WeekStart, &weekCompleted, &weekCancelled, &week.TotalMinutes, &week.AverageRPE,
			&week.TotalDistanceKM, &week.TotalElevationM, &week.AveragePowerW, &week.AverageHeartRate); err != nil {
			return evolution.Summary{}, err
		}
		week.CompletedSessions, week.CancelledSessions = int(weekCompleted), int(weekCancelled)
		result.Weeks = append(result.Weeks, week)
	}
	if err := weeks.Err(); err != nil {
		return evolution.Summary{}, err
	}

	recentRows, err := s.pool.Query(ctx, `
		SELECT ws.completed_at::date::text, w.name, w.duration_minutes,
			COALESCE(w.target_rpe, 0)::double precision,
			ws.duration_minutes, ws.actual_rpe,
			ws.distance_km::double precision, ws.average_power_watts,
			ws.average_heart_rate, f.fatigue_after,
			COALESCE(f.pain_reported, false)
		FROM workout_sessions ws
		JOIN athlete_profiles ap ON ap.id = ws.athlete_profile_id
		JOIN workouts w ON w.id = ws.workout_id
		LEFT JOIN feedback f ON f.workout_session_id = ws.id
		WHERE ap.user_id = $1 AND ws.status = 'completed'
		ORDER BY ws.completed_at DESC, ws.created_at DESC
		LIMIT 12`, userID)
	if err != nil {
		return evolution.Summary{}, err
	}
	defer recentRows.Close()
	for recentRows.Next() {
		var session evolution.SessionComparison
		var actualMinutes, fatigueAfter *int
		var actualRPE *float64
		var distanceKM *float64
		var averagePowerW, averageHeartRate *int
		if err := recentRows.Scan(&session.CompletedOn, &session.Name, &session.PlannedMinutes, &session.PlannedRPE,
			&actualMinutes, &actualRPE, &distanceKM, &averagePowerW, &averageHeartRate, &fatigueAfter, &session.PainReported); err != nil {
			return evolution.Summary{}, err
		}
		session.ActualMinutes, session.ActualRPE = actualMinutes, actualRPE
		session.DistanceKM, session.AveragePowerW = distanceKM, averagePowerW
		session.AverageHeartRate, session.FatigueAfter = averageHeartRate, fatigueAfter
		if actualMinutes != nil {
			delta := *actualMinutes - session.PlannedMinutes
			session.DurationDeltaMinutes = &delta
		}
		if actualRPE != nil {
			delta := *actualRPE - session.PlannedRPE
			session.RPEDelta = &delta
		}
		result.RecentSessions = append(result.RecentSessions, session)
	}
	if err := recentRows.Err(); err != nil {
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
