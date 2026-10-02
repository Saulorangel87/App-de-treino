ALTER TABLE cycling_assessments
    DROP CONSTRAINT cycling_assessments_halves_together,
    DROP COLUMN heart_rate_second_half,
    DROP COLUMN heart_rate_first_half,
    DROP COLUMN distance_km,
    DROP COLUMN average_power_w,
    DROP COLUMN average_heart_rate;
