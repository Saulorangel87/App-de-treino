ALTER TABLE scientific_sources
    ADD COLUMN population_studied TEXT,
    ADD COLUMN research_objective TEXT,
    ADD COLUMN stimulus_analyzed TEXT,
    ADD COLUMN expected_benefits TEXT,
    ADD COLUMN limitations TEXT,
    ADD COLUMN risks TEXT,
    ADD COLUMN contraindications TEXT,
    ADD COLUMN confidence_level TEXT NOT NULL DEFAULT 'not_calibrated'
        CHECK (confidence_level IN ('low', 'moderate', 'high', 'not_calibrated')),
    ADD COLUMN last_reviewed_on DATE,
    ADD COLUMN related_rules TEXT[];

UPDATE scientific_sources
SET
    population_studied = 'Consultar a população e os critérios de inclusão descritos na publicação original.',
    research_objective = summary,
    stimulus_analyzed = training_focus,
    expected_benefits = summary,
    limitations = 'A fonte orienta o contexto da decisão, mas não valida uma dose universal para todos os atletas.',
    risks = 'Riscos específicos não foram calibrados pelo Cadência; prevalecem as travas de dor, limitação e recuperação.',
    contraindications = 'Não usar isoladamente para liberar carga, esforço máximo ou substituir avaliação profissional.',
    last_reviewed_on = DATE '2026-09-14',
    related_rules = ARRAY['rules-v1', training_focus]
WHERE population_studied IS NULL;

ALTER TABLE scientific_sources
    ALTER COLUMN population_studied SET NOT NULL,
    ALTER COLUMN research_objective SET NOT NULL,
    ALTER COLUMN stimulus_analyzed SET NOT NULL,
    ALTER COLUMN expected_benefits SET NOT NULL,
    ALTER COLUMN limitations SET NOT NULL,
    ALTER COLUMN risks SET NOT NULL,
    ALTER COLUMN contraindications SET NOT NULL,
    ALTER COLUMN last_reviewed_on SET NOT NULL,
    ALTER COLUMN related_rules SET NOT NULL,
    ADD CONSTRAINT scientific_sources_related_rules_check CHECK (cardinality(related_rules) > 0);
