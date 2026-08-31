ALTER TABLE workouts DROP CONSTRAINT workouts_status_check;
ALTER TABLE workouts
    ADD CONSTRAINT workouts_status_check
    CHECK (status IN ('planned', 'in_progress', 'completed', 'skipped', 'adapted'));

ALTER TABLE workout_sessions
    ADD COLUMN status TEXT NOT NULL DEFAULT 'in_progress',
    ADD COLUMN cancelled_at TIMESTAMPTZ;

UPDATE workout_sessions
SET status = 'completed'
WHERE completed_at IS NOT NULL;

ALTER TABLE workout_sessions
    ADD CONSTRAINT workout_sessions_status_check
    CHECK (status IN ('in_progress', 'completed', 'cancelled')),
    ADD CONSTRAINT workout_sessions_terminal_time_check
    CHECK (
        (status = 'in_progress' AND completed_at IS NULL AND cancelled_at IS NULL)
        OR (status = 'completed' AND completed_at IS NOT NULL AND cancelled_at IS NULL)
        OR (status = 'cancelled' AND completed_at IS NULL AND cancelled_at IS NOT NULL)
    );

CREATE UNIQUE INDEX idx_sessions_one_in_progress_per_workout
    ON workout_sessions (workout_id)
    WHERE status = 'in_progress';
