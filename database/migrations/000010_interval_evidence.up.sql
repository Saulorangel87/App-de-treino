INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES (
    'rosenblat-2020',
    'Effect of High-Intensity Interval Training Versus Sprint Interval Training on Time-Trial Performance',
    'Rosenblat, Perrotta e Thomas',
    2020,
    'https://pubmed.ncbi.nlm.nih.gov/32034701/',
    'controlled_intervals',
    'systematic_review',
    'Revisão e meta-análise que diferencia intervalos intensos controlados de sprints e relata resultados favoráveis para blocos mais longos em pessoas ativas e treinadas.'
)
ON CONFLICT (source_key) DO UPDATE SET title = EXCLUDED.title, authors = EXCLUDED.authors, published_year = EXCLUDED.published_year, url = EXCLUDED.url, training_focus = EXCLUDED.training_focus, evidence_level = EXCLUDED.evidence_level, summary = EXCLUDED.summary;
