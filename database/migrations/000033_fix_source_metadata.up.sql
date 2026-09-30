-- Corrige metadados de referências conferidos no PubMed em 30/09/2026. As chaves
-- (source_key) não mudam, porque já estão gravadas nas explicações dos treinos.
--
--  * acsm-1998: o título cadastrado era o de outro documento (treino de força, 2002);
--    o link e o ano já eram do posicionamento de 1998 sobre quantidade e qualidade
--    do exercício.
--  * bourdon-2017: o link apontava para a revisão de Foster et al. (2017), não para
--    o consenso de Bourdon et al. (2017).
--  * impellizzeri-2020: autor, ano e título não correspondiam ao link; o artigo é de
--    Foster et al. (2021), uma perspectiva histórica, e não uma revisão sistemática.
--  * haddad-2017: é uma revisão da literatura, não uma revisão sistemática.
ALTER TABLE scientific_sources DROP CONSTRAINT IF EXISTS scientific_sources_evidence_level_check;
ALTER TABLE scientific_sources ADD CONSTRAINT scientific_sources_evidence_level_check
    CHECK (evidence_level IN ('primary_study', 'narrative_review', 'systematic_review', 'consensus', 'guideline'));

UPDATE scientific_sources SET
    title = 'American College of Sports Medicine Position Stand. The recommended quantity and quality of exercise for developing and maintaining cardiorespiratory and muscular fitness, and flexibility in healthy adults',
    summary = 'Posicionamento do ACSM sobre a quantidade e a qualidade do exercício (frequência, intensidade e duração) que produz efeito de treino. Apoia o princípio de sobrecarga e progressão gradual de forma geral; não define percentuais universais nem critérios de proteção por dor ou fadiga.',
    research_objective = 'Orientar a quantidade e a qualidade do exercício para desenvolver e manter a aptidão cardiorrespiratória e muscular e a flexibilidade.',
    expected_benefits = 'Apoia o princípio de progressão gradual; não define percentuais universais.',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'acsm-1998';

UPDATE scientific_sources SET
    title = 'Monitoring Athlete Training Loads: Consensus Statement',
    authors = 'Bourdon et al.',
    url = 'https://pubmed.ncbi.nlm.nih.gov/28463642/',
    summary = 'Declaração de consenso sobre o monitoramento da carga de treino de atletas, em treino e competição, para quantificar as cargas interna e externa e ajudar a proteger contra lesão e problemas de saúde. Não define limiares numéricos de dor ou fadiga.',
    research_objective = 'Reunir recomendações de especialistas sobre como monitorar e interpretar a carga de treino de atletas.',
    expected_benefits = 'Apoia monitorar a carga do atleta e interpretá-la no contexto; não fornece limiares numéricos.',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'bourdon-2017';

UPDATE scientific_sources SET
    title = '25 Years of Session Rating of Perceived Exertion: Historical Perspective and Development',
    authors = 'Foster et al.',
    published_year = 2021,
    evidence_level = 'narrative_review',
    summary = 'Perspectiva histórica sobre o desenvolvimento e o uso do session-RPE ao longo de 25 anos.',
    research_objective = 'Descrever a origem, o desenvolvimento e o uso do método session-RPE.',
    expected_benefits = 'Apoia o uso do esforço percebido da sessão para acompanhar a carga interna.',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'impellizzeri-2020';

UPDATE scientific_sources SET
    evidence_level = 'narrative_review',
    summary = 'Revisão da literatura sobre a validade, a utilidade ecológica e os fatores que influenciam o session-RPE para monitorar a carga de treino.',
    research_objective = 'Reunir dados de validade do session-RPE, o racional do método e os fatores que podem alterar o esforço percebido.',
    expected_benefits = 'Apoia o uso do session-RPE para monitorar a carga de treino.',
    last_reviewed_on = DATE '2026-09-30'
WHERE source_key = 'haddad-2017';
