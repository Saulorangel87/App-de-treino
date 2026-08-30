CREATE UNIQUE INDEX idx_training_plans_one_active_per_athlete
    ON training_plans (athlete_profile_id)
    WHERE status = 'active';
