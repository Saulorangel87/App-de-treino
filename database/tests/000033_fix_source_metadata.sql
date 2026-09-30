-- Metadados das referências corrigidos na 000033 (conferidos no PubMed em 30/09/2026).
BEGIN;

DO $$
DECLARE
    n INTEGER;
BEGIN
    -- O consenso de Bourdon et al. (2017) é o PMID 28463642; o link antigo era da revisão de Foster et al.
    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key = 'bourdon-2017' AND url = 'https://pubmed.ncbi.nlm.nih.gov/28463642/'
      AND title = 'Monitoring Athlete Training Loads: Consensus Statement' AND evidence_level = 'consensus';
    IF n <> 1 THEN RAISE EXCEPTION 'bourdon-2017 must point to the consensus statement'; END IF;

    -- O artigo do PMID 33508782 é de Foster et al. (2021), uma perspectiva histórica.
    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key = 'impellizzeri-2020' AND authors = 'Foster et al.' AND published_year = 2021
      AND evidence_level = 'narrative_review' AND title LIKE '25 Years of Session Rating of Perceived Exertion%';
    IF n <> 1 THEN RAISE EXCEPTION 'impellizzeri-2020 must describe the Foster et al. (2021) article'; END IF;

    -- Revisão da literatura, não sistemática.
    SELECT count(*) INTO n FROM scientific_sources WHERE source_key = 'haddad-2017' AND evidence_level = 'narrative_review';
    IF n <> 1 THEN RAISE EXCEPTION 'haddad-2017 is a narrative review'; END IF;

    -- O título do ACSM 1998 não pode ser o de treino de força.
    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key = 'acsm-1998' AND title LIKE '%recommended quantity and quality of exercise%'
      AND title NOT LIKE '%Resistance Training%';
    IF n <> 1 THEN RAISE EXCEPTION 'acsm-1998 title must match the 1998 position stand'; END IF;

    -- Nenhuma referência pode continuar sem revisão na data da correção entre as quatro corrigidas.
    SELECT count(*) INTO n FROM scientific_sources
    WHERE source_key IN ('acsm-1998', 'bourdon-2017', 'impellizzeri-2020', 'haddad-2017')
      AND last_reviewed_on <> DATE '2026-09-30';
    IF n <> 0 THEN RAISE EXCEPTION 'corrected sources must record the review date'; END IF;

    -- O novo nível continua sujeito à restrição: valores inventados são recusados.
    BEGIN
        UPDATE scientific_sources SET evidence_level = 'opinion' WHERE source_key = 'foster-2001';
        RAISE EXCEPTION 'an unknown evidence level should fail';
    EXCEPTION WHEN check_violation THEN
        NULL;
    END;
END;
$$;

ROLLBACK;
