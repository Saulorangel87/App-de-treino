ALTER TABLE scientific_sources
    DROP COLUMN IF EXISTS related_rules,
    DROP COLUMN IF EXISTS last_reviewed_on,
    DROP COLUMN IF EXISTS confidence_level,
    DROP COLUMN IF EXISTS contraindications,
    DROP COLUMN IF EXISTS risks,
    DROP COLUMN IF EXISTS limitations,
    DROP COLUMN IF EXISTS expected_benefits,
    DROP COLUMN IF EXISTS stimulus_analyzed,
    DROP COLUMN IF EXISTS research_objective,
    DROP COLUMN IF EXISTS population_studied;
