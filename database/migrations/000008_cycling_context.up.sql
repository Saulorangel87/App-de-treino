ALTER TABLE athlete_profiles
    ADD COLUMN cycling_context JSONB NOT NULL DEFAULT '{}'::jsonb;
