ALTER TABLE injuries_or_limitations
    DROP CONSTRAINT IF EXISTS injuries_or_limitations_symptoms_check,
    DROP COLUMN IF EXISTS medical_restriction,
    DROP COLUMN IF EXISTS symptoms_during_after;
