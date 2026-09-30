-- Auditoria das demais referências contra o PubMed (30/09/2026). As 27 fontes do
-- catálogo foram conferidas (título, autores, ano, link e tipo de publicação);
-- corrige as três divergências restantes. As chaves não mudam.
--
--  * post-competition-recovery-2019: revisão narrativa (sem busca sistemática
--    descrita), não revisão sistemática.
--  * haddad-2017: título traduzido; passa ao título original, como as demais.
--  * short-self-paced-2025: título abreviado e link só da editora; passa ao título
--    completo e ao registro do PubMed (PMID 39973903).
UPDATE scientific_sources SET
    evidence_level = 'narrative_review',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'post-competition-recovery-2019';

UPDATE scientific_sources SET
    title = 'Session-RPE Method for Training Load Monitoring: Validity, Ecological Usefulness, and Influencing Factors',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'haddad-2017';

UPDATE scientific_sources SET
    title = 'Effect of self-paced sprint interval training and low-volume HIIT on cardiorespiratory fitness: the role of heart rate and power output',
    url = 'https://pubmed.ncbi.nlm.nih.gov/39973903/',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'short-self-paced-2025';
