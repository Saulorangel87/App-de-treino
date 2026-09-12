INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES
(
    'road-vo2-intervention-2024',
    'Greater improvement in aerobic capacity after a polarized training program including cycling interval training at low cadence (50-70 RPM) than freely chosen cadence (above 80 RPM)',
    'Rafal Hebisz e Paulina Hebisz',
    2024,
    'https://pubmed.ncbi.nlm.nih.gov/39536034/',
    'road_vo2_interval_training',
    'primary_study',
    'Ensaio com 24 ciclistas mulheres bem treinadas: programas polarizados de oito semanas incluíram blocos de 4 minutos a 90-100% da potência aeróbica máxima. O resultado informa o formato e a população estudada, mas não sustenta copiar a carga ou o alvo de potência para todos os perfis; o Cadência usa somente uma adaptação conservadora guiada por RPE.'
),
(
    'road-vo2-response-2024',
    'The higher the fraction of maximal oxygen uptake is during interval training, the greater is the cycling performance gain',
    'Ingvill Odden et al.',
    2024,
    'https://pubmed.ncbi.nlm.nih.gov/39385317/',
    'road_vo2_interval_training',
    'primary_study',
    'Estudo observacional de intervenção com 22 ciclistas bem treinados durante nove semanas: maior fração de VO2max nos intervalos de 8 minutos se associou a maiores ganhos em indicadores de desempenho. A associação não define uma dose universal nem autoriza usar frequência cardíaca ou potência como substitutos perfeitos de VO2max.'
)
ON CONFLICT (source_key) DO UPDATE SET
    title = EXCLUDED.title,
    authors = EXCLUDED.authors,
    published_year = EXCLUDED.published_year,
    url = EXCLUDED.url,
    training_focus = EXCLUDED.training_focus,
    evidence_level = EXCLUDED.evidence_level,
    summary = EXCLUDED.summary;
