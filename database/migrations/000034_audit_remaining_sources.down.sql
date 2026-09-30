UPDATE scientific_sources SET evidence_level = 'systematic_review'
WHERE source_key = 'post-competition-recovery-2019';

UPDATE scientific_sources SET title = 'Validade do método session-RPE para quantificar carga de treino'
WHERE source_key = 'haddad-2017';

UPDATE scientific_sources SET
    title = 'Effect of self-paced sprint interval training and low-volume high-intensity interval training on cardiorespiratory fitness',
    url = 'https://www.frontiersin.org/journals/physiology/articles/10.3389/fphys.2025.1484722/full'
WHERE source_key = 'short-self-paced-2025';
