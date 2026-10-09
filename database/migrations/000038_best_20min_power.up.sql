-- Melhor potência média em 20 minutos seguidos de cada atividade importada.
-- Serve só para sugerir o FTP ao atleta; nunca alimenta o motor de prescrição.
ALTER TABLE imported_activities
    ADD COLUMN best_20min_power_watts INTEGER CHECK (best_20min_power_watts >= 0);
