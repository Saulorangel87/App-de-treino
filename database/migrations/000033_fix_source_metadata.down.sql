-- Restaura os metadados anteriores à correção de 30/09/2026.
UPDATE scientific_sources SET evidence_level = 'systematic_review' WHERE evidence_level = 'narrative_review';
ALTER TABLE scientific_sources DROP CONSTRAINT IF EXISTS scientific_sources_evidence_level_check;
ALTER TABLE scientific_sources ADD CONSTRAINT scientific_sources_evidence_level_check
    CHECK (evidence_level IN ('primary_study', 'systematic_review', 'consensus', 'guideline'));

UPDATE scientific_sources SET
    title = 'Progression Models in Resistance Training for Healthy Adults',
    summary = 'Apoia o princípio de progressão gradual; não define percentuais universais.',
    research_objective = 'Apoia o princípio de progressão gradual; não define percentuais universais.',
    last_reviewed_on = DATE '2026-09-14'
WHERE source_key = 'acsm-1998';

UPDATE scientific_sources SET
    title = 'Monitoramento de carga em esporte de alto rendimento',
    url = 'https://pubmed.ncbi.nlm.nih.gov/28253038/',
    summary = 'Consenso sobre interpretação contextual de carga.',
    research_objective = 'Consenso sobre interpretação contextual de carga.',
    expected_benefits = 'Consenso sobre interpretação contextual de carga.',
    last_reviewed_on = DATE '2026-09-14'
WHERE source_key = 'bourdon-2017';

UPDATE scientific_sources SET
    title = 'Vinte e cinco anos de monitoramento de carga por session-RPE',
    authors = 'Impellizzeri et al.',
    published_year = 2020,
    summary = 'Revisão sobre session-RPE e suas limitações.',
    research_objective = 'Revisão sobre session-RPE e suas limitações.',
    expected_benefits = 'Revisão sobre session-RPE e suas limitações.',
    last_reviewed_on = DATE '2026-09-14'
WHERE source_key = 'impellizzeri-2020';

UPDATE scientific_sources SET
    summary = 'Revisão sistemática sobre a validade do session-RPE para monitorar carga.',
    research_objective = 'Revisão sistemática sobre a validade do session-RPE para monitorar carga.',
    expected_benefits = 'Revisão sistemática sobre a validade do session-RPE para monitorar carga.',
    last_reviewed_on = DATE '2026-09-14'
WHERE source_key = 'haddad-2017';
