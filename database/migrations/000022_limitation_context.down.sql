ALTER TABLE injuries_or_limitations
    DROP COLUMN IF EXISTS started_on,
    DROP COLUMN IF EXISTS aggravating_movement,
    DROP COLUMN IF EXISTS pain_intensity,
    DROP COLUMN IF EXISTS location;
