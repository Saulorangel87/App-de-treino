ALTER TABLE athlete_profiles
    ADD COLUMN waist_cm NUMERIC(5,2) CHECK (waist_cm BETWEEN 30 AND 250),
    ADD COLUMN body_fat_percent NUMERIC(4,1) CHECK (body_fat_percent BETWEEN 2 AND 70),
    ADD COLUMN weight_trend TEXT NOT NULL DEFAULT 'not_informed'
        CHECK (weight_trend IN ('not_informed', 'stable', 'increasing', 'decreasing'));

ALTER TABLE injuries_or_limitations
    ADD COLUMN recent_surgery BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN exercise_prohibited BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN condition_affecting_exercise BOOLEAN NOT NULL DEFAULT false;
