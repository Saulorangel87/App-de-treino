ALTER TABLE injuries_or_limitations
    DROP COLUMN condition_affecting_exercise,
    DROP COLUMN exercise_prohibited,
    DROP COLUMN recent_surgery;

ALTER TABLE athlete_profiles
    DROP COLUMN weight_trend,
    DROP COLUMN body_fat_percent,
    DROP COLUMN waist_cm;
