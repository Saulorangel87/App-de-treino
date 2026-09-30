-- Execução importada anexada ao plano (backend/internal/repository/imported_execution.go).
-- A consulta abaixo repete a do repositório: se ela mudar, esta fixture muda junto.
BEGIN;

DO $$
DECLARE
    user_a UUID; user_b UUID; profile_a UUID; plan_a UUID; w1 UUID; w2 UUID;
    n INTEGER; chosen INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name) VALUES ('exec-a@example.invalid', 'test-only', 'A') RETURNING id INTO user_a;
    INSERT INTO users (email, password_hash, display_name) VALUES ('exec-b@example.invalid', 'test-only', 'B') RETURNING id INTO user_b;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_a, 'intermediate') RETURNING id INTO profile_a;
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status) VALUES (profile_a, DATE '2099-08-03', DATE '2099-08-30', 'active') RETURNING id INTO plan_a;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-04', 'Um', 'T', 60, 4, '{}'::jsonb, '{}'::jsonb, 'completed') RETURNING id INTO w1;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-06', 'Dois', 'T', 60, 4, '{}'::jsonb, '{}'::jsonb, 'planned') RETURNING id INTO w2;

    -- Duas atividades do atleta no mesmo treino: vale a importada por último.
    INSERT INTO imported_activities (user_id, workout_id, source, file_hash, started_at, moving_seconds, imported_at)
    VALUES (user_a, w1, 'fit', repeat('1', 64), now(), 3000, now() - interval '1 hour');
    INSERT INTO imported_activities (user_id, workout_id, source, file_hash, started_at, moving_seconds, imported_at)
    VALUES (user_a, w1, 'fit', repeat('2', 64), now(), 3600, now());
    -- Atividade de outro usuário apontando para o treino do atleta A nunca aparece.
    INSERT INTO imported_activities (user_id, workout_id, source, file_hash, started_at, moving_seconds)
    VALUES (user_b, w2, 'gpx', repeat('3', 64), now(), 1200);

    SELECT count(*) INTO n FROM (
        SELECT DISTINCT ON (ia.workout_id) ia.workout_id, ia.moving_seconds
        FROM imported_activities ia JOIN workouts w ON w.id = ia.workout_id
        WHERE ia.user_id = user_a AND w.training_plan_id = plan_a
        ORDER BY ia.workout_id, ia.imported_at DESC
    ) linked;
    IF n <> 1 THEN RAISE EXCEPTION 'only the athlete''s own linked workout must be returned, got %', n; END IF;

    SELECT moving_seconds INTO chosen FROM (
        SELECT DISTINCT ON (ia.workout_id) ia.workout_id, ia.moving_seconds
        FROM imported_activities ia JOIN workouts w ON w.id = ia.workout_id
        WHERE ia.user_id = user_a AND w.training_plan_id = plan_a
        ORDER BY ia.workout_id, ia.imported_at DESC
    ) linked;
    IF chosen <> 3600 THEN RAISE EXCEPTION 'the most recently imported activity must win, got %', chosen; END IF;
END;
$$;

ROLLBACK;
