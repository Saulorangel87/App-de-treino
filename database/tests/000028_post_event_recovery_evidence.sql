BEGIN;

DO $$
DECLARE
    source_count INTEGER;
    source_total INTEGER;
    source_confidence TEXT;
    source_rule TEXT;
BEGIN
    SELECT COUNT(*), MIN(confidence_level), MIN(related_rules[1])
    INTO source_count, source_confidence, source_rule
    FROM scientific_sources
    WHERE source_key = 'post-competition-recovery-2019';

    SELECT COUNT(*) INTO source_total
    FROM scientific_sources
    WHERE source_key IN ('post-competition-recovery-2019', 'recovery-umbrella-2024');

    IF source_count <> 1 OR source_total <> 2 OR source_confidence <> 'moderate' OR source_rule <> 'rules-v1' THEN
        RAISE EXCEPTION 'post-event recovery evidence metadata is missing or invalid';
    END IF;
END;
$$;

ROLLBACK;
