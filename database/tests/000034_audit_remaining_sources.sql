-- Correções da auditoria das demais referências (000034).
BEGIN;

DO $$
DECLARE
    n INTEGER;
BEGIN
    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key = 'post-competition-recovery-2019' AND evidence_level = 'narrative_review';
    IF n <> 1 THEN RAISE EXCEPTION 'post-competition-recovery-2019 is a narrative review'; END IF;

    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key = 'haddad-2017' AND title LIKE 'Session-RPE Method for Training Load Monitoring%';
    IF n <> 1 THEN RAISE EXCEPTION 'haddad-2017 must use the original title'; END IF;

    -- Toda referência do catálogo aponta para o registro do PubMed.
    SELECT count(*) INTO n FROM scientific_sources WHERE url NOT LIKE 'https://pubmed.ncbi.nlm.nih.gov/%';
    IF n <> 0 THEN RAISE EXCEPTION 'every source must link to its PubMed record, % do not', n; END IF;
END;
$$;

ROLLBACK;
