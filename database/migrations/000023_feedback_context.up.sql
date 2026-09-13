ALTER TABLE feedback
    ADD COLUMN satisfaction SMALLINT,
    ADD COLUMN terrain TEXT,
    ADD COLUMN external_conditions TEXT,
    ADD CONSTRAINT feedback_satisfaction_check
        CHECK (satisfaction IS NULL OR satisfaction BETWEEN 1 AND 5),
    ADD CONSTRAINT feedback_terrain_check
        CHECK (terrain IS NULL OR terrain IN ('flat', 'rolling', 'hilly', 'mixed', 'technical', 'indoor')),
    ADD CONSTRAINT feedback_external_conditions_check
        CHECK (external_conditions IS NULL OR external_conditions IN ('normal', 'heat', 'cold', 'wind', 'rain', 'poor_visibility', 'other'));
