INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES
(
    'taper-meta-2023',
    'Effects of tapering on performance in endurance athletes: A systematic review and meta-analysis',
    'Zhiqiang Wang, Yong Wang, Weifeng Gao e Yaping Zhong',
    2023,
    'https://pubmed.ncbi.nlm.nih.gov/37163550/',
    'event_taper',
    'systematic_review',
    'Meta-análise de 14 estudos: o taper melhorou tempo de prova e tempo até a exaustão; a síntese sugere redução progressiva de volume por até 21 dias, mantendo intensidade e frequência. A evidência combina modalidades de endurance e não define uma dose universal.'
),
(
    'taper-cyclist-2025',
    'Unaltered maximal power and submaximal performance correlates with an oxidative vastus lateralis proteome phenotype during tapering in male cyclists',
    'Pieter de Lange et al.',
    2025,
    'https://pubmed.ncbi.nlm.nih.gov/40265526/',
    'event_taper',
    'primary_study',
    'Estudo com ciclistas homens bem treinados: duas semanas com redução aproximada de 50% do volume e intensidade mantida preservaram a maior parte das adaptações de desempenho. A população, o período de treino e a resposta submáxima limitam a transferência para outros perfis.'
),
(
    'taper-overreach-cyclists-2023',
    'Effect of intensified training on cognitive function, psychological state & performance in trained cyclists',
    'Sarah E. Costello et al.',
    2023,
    'https://pubmed.ncbi.nlm.nih.gov/35771645/',
    'event_taper_safety',
    'primary_study',
    'Ensaio com ciclistas treinados: duas semanas de intensificação aumentaram carga e pioraram temporariamente desempenho, humor e equilíbrio recuperação-estresse; após duas semanas de taper, as medidas retornaram à linha de base. Não sustenta sobrecarga automática no Cadência.'
)
ON CONFLICT (source_key) DO UPDATE SET
    title = EXCLUDED.title,
    authors = EXCLUDED.authors,
    published_year = EXCLUDED.published_year,
    url = EXCLUDED.url,
    training_focus = EXCLUDED.training_focus,
    evidence_level = EXCLUDED.evidence_level,
    summary = EXCLUDED.summary;
