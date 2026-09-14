ALTER TABLE workout_sessions
    ADD COLUMN average_cadence_rpm INTEGER,
    ADD CONSTRAINT workout_sessions_average_cadence_rpm_check
        CHECK (average_cadence_rpm IS NULL OR average_cadence_rpm BETWEEN 1 AND 300);
