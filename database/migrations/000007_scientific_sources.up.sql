CREATE TABLE scientific_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_key TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    authors TEXT NOT NULL,
    published_year SMALLINT NOT NULL CHECK (published_year BETWEEN 1900 AND 2100),
    url TEXT NOT NULL,
    training_focus TEXT NOT NULL,
    evidence_level TEXT NOT NULL CHECK (evidence_level IN ('primary_study', 'systematic_review', 'consensus', 'guideline')),
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary) VALUES
('foster-2001', 'A new approach to monitoring exercise training', 'Foster et al.', 2001, 'https://pubmed.ncbi.nlm.nih.gov/11708692/', 'load_monitoring', 'primary_study', 'Fundamenta o uso do esforço percebido da sessão para acompanhar carga interna.'),
('haddad-2017', 'Validade do método session-RPE para quantificar carga de treino', 'Haddad et al.', 2017, 'https://pubmed.ncbi.nlm.nih.gov/29163016/', 'load_monitoring', 'systematic_review', 'Revisão sistemática sobre a validade do session-RPE para monitorar carga.'),
('impellizzeri-2020', 'Vinte e cinco anos de monitoramento de carga por session-RPE', 'Impellizzeri et al.', 2020, 'https://pubmed.ncbi.nlm.nih.gov/33508782/', 'load_monitoring', 'systematic_review', 'Revisão sobre session-RPE e suas limitações.'),
('bourdon-2017', 'Monitoramento de carga em esporte de alto rendimento', 'Bourdon et al.', 2017, 'https://pubmed.ncbi.nlm.nih.gov/28253038/', 'load_monitoring', 'consensus', 'Consenso sobre interpretação contextual de carga.'),
('acsm-1998', 'Progression Models in Resistance Training for Healthy Adults', 'American College of Sports Medicine', 1998, 'https://pubmed.ncbi.nlm.nih.gov/9624661/', 'aerobic_progression', 'guideline', 'Apoia o princípio de progressão gradual; não define percentuais universais.')
ON CONFLICT (source_key) DO UPDATE SET title = EXCLUDED.title, authors = EXCLUDED.authors, published_year = EXCLUDED.published_year, url = EXCLUDED.url, training_focus = EXCLUDED.training_focus, evidence_level = EXCLUDED.evidence_level, summary = EXCLUDED.summary;
