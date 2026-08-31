DROP INDEX IF EXISTS idx_sessions_one_in_progress_per_workout;

UPDATE workouts
SET status = 'planned'
WHERE status = 'in_progress';

ALTER TABLE workout_sessions
    DROP CONSTRAINT IF EXISTS workout_sessions_terminal_time_check,
    DROP CONSTRAINT IF EXISTS workout_sessions_status_check,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS status;

ALTER TABLE workouts DROP CONSTRAINT workouts_status_check;
ALTER TABLE workouts
    ADD CONSTRAINT workouts_status_check
    CHECK (status IN ('planned', 'completed', 'skipped', 'adapted'));
