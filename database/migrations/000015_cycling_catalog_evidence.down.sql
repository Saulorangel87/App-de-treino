DELETE FROM scientific_sources
WHERE source_key IN (
    'road-intensity-2024',
    'road-mit-block-2025',
    'road-block-comparison-2025',
    'road-strength-2026',
    'xco-physiology-2026',
    'xco-power-distribution-2021',
    'xco-pacing-2021',
    'gravel-field-2024',
    'dh-injury-2024',
    'mtb-crash-mechanisms-2025',
    'track-sprint-load-2023'
);
