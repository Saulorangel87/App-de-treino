INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES
(
    'short-self-paced-2025',
    'Effect of self-paced sprint interval training and low-volume high-intensity interval training on cardiorespiratory fitness',
    'Hesketh et al.',
    2025,
    'https://www.frontiersin.org/journals/physiology/articles/10.3389/fphys.2025.1484722/full',
    'short_self_regulated_intervals',
    'primary_study',
    'Ensaio randomizado cruzado de seis semanas em 82 adultos anteriormente inativos: comparou 4–8 repetições de 30 segundos com 120 segundos de recuperação e 6–10 repetições de 1 minuto com 1 minuto de recuperação. Ambos os formatos melhoraram o VO₂peak, mas a população não representa automaticamente ciclistas treinados e não define uma dose universal.'
),
(
    'short-interval-cyclists-2020',
    'Superior performance improvements in elite cyclists following short-interval vs effort-matched long-interval training',
    'Rønnestad et al.',
    2020,
    'https://pubmed.ncbi.nlm.nih.gov/31977120/',
    'short_self_regulated_intervals',
    'primary_study',
    'Ensaio com 18 ciclistas de elite que comparou blocos curtos repetidos de 30 segundos com intervalos longos de 5 minutos durante três semanas. Alguns indicadores favoreceram o formato curto, mas o nível de treinamento, a amostra pequena e o esforço intenso limitam a transferência para perfis gerais.'
)
ON CONFLICT (source_key) DO UPDATE SET
    title = EXCLUDED.title,
    authors = EXCLUDED.authors,
    published_year = EXCLUDED.published_year,
    url = EXCLUDED.url,
    training_focus = EXCLUDED.training_focus,
    evidence_level = EXCLUDED.evidence_level,
    summary = EXCLUDED.summary;
