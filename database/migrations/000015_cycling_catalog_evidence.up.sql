INSERT INTO scientific_sources (source_key, title, authors, published_year, url, training_focus, evidence_level, summary)
VALUES
(
    'road-intensity-2024',
    'Comparison of Polarized Versus Other Types of Endurance Training Intensity Distribution on Athletes'' Endurance Performance: A Systematic Review with Meta-analysis',
    'Pedro Silva Oliveira, Giorjines Boppre e Hélder Fonseca',
    2024,
    'https://pubmed.ncbi.nlm.nih.gov/38717713/',
    'road_intensity_distribution',
    'systematic_review',
    'Revisão e meta-análise de 17 estudos: o treinamento polarizado favoreceu VO2peak em alguns subgrupos, mas os demais desfechos de resistência foram semelhantes entre distribuições. Não define uma receita universal.'
),
(
    'road-mit-block-2025',
    'A Moderate-Intensity Interval Training Block Improves Endurance Performance in Well-Trained Cyclists',
    'Knut Sindre Mølmen et al.',
    2025,
    'https://pubmed.ncbi.nlm.nih.gov/40101160/',
    'road_interval_training',
    'primary_study',
    'Ensaio cruzado com 30 ciclistas homens bem treinados; seis sessões intervaladas moderadas em sete dias, seguidas de recuperação ativa, melhoraram alguns indicadores. A população e o bloco não devem ser aplicados automaticamente a iniciantes.'
),
(
    'road-block-comparison-2025',
    'Block Training With Moderate- or High-Intensity Intervals Both Improve Endurance Performance in Well-Trained Cyclists',
    'Bent R. Rønnestad et al.',
    2025,
    'https://pubmed.ncbi.nlm.nih.gov/41169000/',
    'road_interval_training',
    'primary_study',
    'Comparação de blocos moderados e intensos em 22 ciclistas bem treinados; ambos melhoraram alguns indicadores, com respostas específicas à intensidade. Não sustenta blocos concentrados para perfis sem elegibilidade.'
),
(
    'road-strength-2026',
    'Heavy strength training effects on physiological determinants of endurance cyclist performance: a systematic review with meta-analysis',
    'Cristian Llanos-Lagos, Rodrigo Ramirez-Campillo e Eduardo Sáez de Villarreal',
    2026,
    'https://pubmed.ncbi.nlm.nih.gov/40632222/',
    'strength_training',
    'systematic_review',
    'Revisão e meta-análise de 17 estudos com 262 participantes: força pesada favoreceu eficiência, potência anaeróbica e desempenho, mas a certeza foi baixa e não define a implementação ótima.'
),
(
    'xco-physiology-2026',
    'The Physiology of Contemporary Olympic Cross-Country Mountain Biking: A Systematic Review',
    'Gabriel Protzen et al.',
    2026,
    'https://pubmed.ncbi.nlm.nih.gov/41739301/',
    'xco_demands',
    'systematic_review',
    'Revisão de 53 estudos sobre XCO contemporâneo: alta demanda aeróbica, maior relevância de esforços curtos intensos e padrão intermitente. Intervenções com avaliação direta de desempenho ainda são escassas.'
),
(
    'xco-power-distribution-2021',
    'Aerobic and Anaerobic Power Distribution During Cross-Country Mountain Bike Racing',
    'Bernhard Prinz, Dieter Simon, Harald Tschan e Alfred Nimmerichter',
    2021,
    'https://pubmed.ncbi.nlm.nih.gov/33848975/',
    'xco_demands',
    'primary_study',
    'Estudo de demanda com ciclistas de elite em 119 provas: cerca de 30% do tempo acima da potência aeróbica máxima e centenas de esforços curtos. Descreve a prova, não valida sozinho um protocolo.'
),
(
    'xco-pacing-2021',
    'Exercise Intensity and Pacing Pattern During a Cross-Country Olympic Mountain Bike Race',
    'Steffan Næss et al.',
    2021,
    'https://pubmed.ncbi.nlm.nih.gov/34349670/',
    'xco_demands',
    'primary_study',
    'Estudo de pacing em sete ciclistas competitivos de XCO: numerosas ações curtas acima da potência crítica e redução da magnitude ao longo das voltas. Usado para especificidade, não como prescrição universal.'
),
(
    'gravel-field-2024',
    'Fluid Intake and Hydration Responses to Mass Participation Gravel Cycling',
    'Carly Schuerger et al.',
    2024,
    'https://pubmed.ncbi.nlm.nih.gov/39807388/',
    'gravel_context',
    'primary_study',
    'Estudo de campo com 121 participantes em uma prova gravel: distância influenciou perdas de massa e ingestão de líquidos, carboidratos e sódio. Informa contexto de prova, não um treino automatizado.'
),
(
    'dh-injury-2024',
    'Downhill race for a rainbow jersey: the epidemiology of injuries in downhill mountain biking at the 2023 UCI cycling world championships-a prospective cohort study of 230 elite cyclists',
    'Thomas Fallon et al.',
    2024,
    'https://pubmed.ncbi.nlm.nih.gov/39411021/',
    'mtb_safety',
    'primary_study',
    'Coorte prospectiva de 230 ciclistas de elite no downhill: observou incidência relevante de lesões em treinamento e competição. Serve como evidência de segurança e não como base para prescrição automática técnica.'
),
(
    'mtb-crash-mechanisms-2025',
    'Injury Mechanisms in Mountain Biking: A Systematic Video Analysis of 534 Cases',
    'S. Bonte et al.',
    2025,
    'https://pubmed.ncbi.nlm.nih.gov/40534393/',
    'mtb_safety',
    'primary_study',
    'Análise de 534 vídeos de quedas em mountain bike: descreve cenários, cinemática e regiões lesionadas. Reforça que habilidade técnica e prevenção não podem ser reduzidas a carga aeróbica.'
),
(
    'track-sprint-load-2023',
    'Training load and intensity distribution for sprinting among world-class track cyclists',
    'François-Denis Desgorces et al.',
    2023,
    'https://pubmed.ncbi.nlm.nih.gov/36961508/',
    'track_sprint_demands',
    'primary_study',
    'Análise retrospectiva de 29 semanas de seis velocistas de pista de nível mundial: predominância de cargas nas zonas mais intensas. Modalidade distinta do catálogo de endurance e mantida fora da primeira liberação.'
)
ON CONFLICT (source_key) DO UPDATE SET
    title = EXCLUDED.title,
    authors = EXCLUDED.authors,
    published_year = EXCLUDED.published_year,
    url = EXCLUDED.url,
    training_focus = EXCLUDED.training_focus,
    evidence_level = EXCLUDED.evidence_level,
    summary = EXCLUDED.summary;
