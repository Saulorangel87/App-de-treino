package repository

import (
	"context"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

// attachImportedExecution adds the measured-vs-planned observation to each workout
// of the plan that has an imported activity linked by the same athlete. When more
// than one activity is linked to a workout, the most recently imported one is used.
// It never changes the prescription: the result lives in Workout.ImportedExecution.
func (s *Store) attachImportedExecution(ctx context.Context, userID string, plan *planning.Plan) error {
	if len(plan.Workouts) == 0 {
		return nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (ia.workout_id)
			ia.workout_id::text, ia.id::text, ia.source, ia.started_at, ia.moving_seconds,
			ia.distance_km::double precision, ia.elevation_gain_m, ia.average_heart_rate, ia.max_heart_rate,
			ia.average_power_watts, ia.normalized_power_watts, ia.average_cadence_rpm
		FROM imported_activities ia
		JOIN workouts w ON w.id = ia.workout_id
		WHERE ia.user_id = $1 AND w.training_plan_id = $2
		ORDER BY ia.workout_id, ia.imported_at DESC`, userID, plan.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	facts := map[string]planning.ImportedActivityFacts{}
	for rows.Next() {
		var workoutID string
		var f planning.ImportedActivityFacts
		var startedAt time.Time
		if err := rows.Scan(&workoutID, &f.ActivityID, &f.Source, &startedAt, &f.MovingSeconds,
			&f.DistanceKM, &f.ElevationGainM, &f.AverageHeartRate, &f.MaxHeartRate,
			&f.AveragePowerW, &f.NormalizedPowerW, &f.AverageCadenceRPM); err != nil {
			return err
		}
		f.StartedAt = startedAt
		facts[workoutID] = f
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for index := range plan.Workouts {
		if f, ok := facts[plan.Workouts[index].ID]; ok {
			assessment := planning.AssessImportedExecution(plan.Workouts[index], f)
			plan.Workouts[index].ImportedExecution = &assessment
		}
	}
	return nil
}
