ALTER TABLE injuries_or_limitations
    ADD COLUMN symptoms_during_after JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN medical_restriction BOOLEAN NOT NULL DEFAULT false,
    ADD CONSTRAINT injuries_or_limitations_symptoms_check
        CHECK (jsonb_typeof(symptoms_during_after) = 'array' AND jsonb_array_length(symptoms_during_after) <= 5);
