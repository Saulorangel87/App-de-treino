BEGIN;

DO $$
DECLARE
    missing_metadata INTEGER;
BEGIN
    SELECT count(*)
    INTO missing_metadata
    FROM scientific_sources
    WHERE population_studied = ''
       OR research_objective = ''
       OR stimulus_analyzed = ''
       OR expected_benefits = ''
       OR limitations = ''
       OR risks = ''
       OR contraindications = ''
       OR last_reviewed_on IS NULL
       OR cardinality(related_rules) = 0;

    IF missing_metadata <> 0 THEN
        RAISE EXCEPTION 'scientific sources without required metadata: %', missing_metadata;
    END IF;

    BEGIN
        INSERT INTO scientific_sources (
            source_key, title, authors, published_year, url, training_focus,
            evidence_level, summary, population_studied, research_objective,
            stimulus_analyzed, expected_benefits, limitations, risks,
            contraindications, confidence_level, last_reviewed_on, related_rules
        ) VALUES (
            'invalid-confidence-test', 'Test', 'Test', 2026,
            'https://example.invalid', 'test', 'primary_study', 'Test',
            'Test', 'Test', 'Test', 'Test', 'Test', 'Test', 'Test',
            'certain', DATE '2026-09-14', ARRAY['rules-v1']
        );
        RAISE EXCEPTION 'invalid confidence level should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
