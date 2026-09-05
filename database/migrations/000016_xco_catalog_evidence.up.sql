INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES (
    'xco-hit-2016',
    'Effects of Sprint versus High-Intensity Aerobic Interval Training on Cross-Country Mountain Biking Performance: A Randomized Controlled Trial',
    'Allan Inoue et al.',
    2016,
    'https://pubmed.ncbi.nlm.nih.gov/26789124/',
    'xco_interval_training',
    'primary_study',
    'Ensaio randomizado com 16 mountain bikers treinados: seis semanas comparando HIT e SIT, com melhora do desempenho de MTB nos dois grupos e vantagem provável do HIT. A população treinada e a carga do estudo não sustentam aplicação universal; o Cadência usa somente uma adaptação aeróbica mais conservadora.'
)
ON CONFLICT (source_key) DO UPDATE SET
    title = EXCLUDED.title,
    authors = EXCLUDED.authors,
    published_year = EXCLUDED.published_year,
    url = EXCLUDED.url,
    training_focus = EXCLUDED.training_focus,
    evidence_level = EXCLUDED.evidence_level,
    summary = EXCLUDED.summary;
