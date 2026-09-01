ALTER TABLE workout_sessions
    ADD COLUMN elevation_gain_m INTEGER CHECK (elevation_gain_m >= 0);
