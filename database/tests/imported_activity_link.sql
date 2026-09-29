-- Vínculo de atividade importada a um treino (backend/internal/repository/
-- imported_activities.go, LinkActivity). O UPDATE abaixo repete a instrução
-- do repositório: se ela mudar, esta fixture precisa mudar junto.
BEGIN;

DO $$
DECLARE
    user_a UUID;
    user_b UUID;
    profile_a UUID;
    profile_b UUID;
    plan_a UUID;
    plan_b UUID;
    workout_a UUID;
    workout_b UUID;
    activity_a UUID;
    activity_b UUID;
    affected BIGINT;
    linked_name TEXT;
    linked_on TEXT;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('link-test-a@example.invalid', 'test-only', 'A') RETURNING id INTO user_a;
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('link-test-b@example.invalid', 'test-only', 'B') RETURNING id INTO user_b;

    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_a, 'intermediate') RETURNING id INTO profile_a;
    INSERT INTO athlete_profiles (user_id, experience_level) VALUES (user_b, 'intermediate') RETURNING id INTO profile_b;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (profile_a, DATE '2099-08-01', DATE '2099-08-28', 'active') RETURNING id INTO plan_a;
    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (profile_b, DATE '2099-08-01', DATE '2099-08-28', 'active') RETURNING id INTO plan_b;

    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_a, DATE '2099-08-10', 'Giro leve protegido', 'Teste', 60, 3, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned')
    RETURNING id INTO workout_a;
    INSERT INTO workouts (training_plan_id, scheduled_on, name, objective, duration_minutes, target_rpe, structure, explanation, status)
    VALUES (plan_b, DATE '2099-08-10', 'Treino do outro atleta', 'Teste', 60, 3, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned')
    RETURNING id INTO workout_b;

    INSERT INTO imported_activities (user_id, source, file_hash, started_at, moving_seconds)
    VALUES (user_a, 'fit', repeat('c', 64), now(), 3600) RETURNING id INTO activity_a;
    INSERT INTO imported_activities (user_id, source, file_hash, started_at, moving_seconds)
    VALUES (user_b, 'fit', repeat('d', 64), now(), 3600) RETURNING id INTO activity_b;

    -- 1) O atleta vincula a própria atividade a um treino do próprio plano.
    UPDATE imported_activities SET workout_id = workout_a
    WHERE id = activity_a AND user_id = user_a
      AND EXISTS (
        SELECT 1 FROM workouts w
        JOIN training_plans tp ON tp.id = w.training_plan_id
        JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
        WHERE w.id = workout_a AND ap.user_id = user_a AND tp.status IN ('active', 'draft', 'completed'));
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 1 THEN
        RAISE EXCEPTION 'linking an own activity to an own workout should update 1 row, got %', affected;
    END IF;

    -- O LEFT JOIN da listagem traz o nome e a data do treino vinculado.
    SELECT w.name, w.scheduled_on::text INTO linked_name, linked_on
    FROM imported_activities ia LEFT JOIN workouts w ON w.id = ia.workout_id
    WHERE ia.id = activity_a AND ia.user_id = user_a;
    IF linked_name <> 'Giro leve protegido' OR linked_on <> '2099-08-10' THEN
        RAISE EXCEPTION 'linked workout projection is wrong: % / %', linked_name, linked_on;
    END IF;

    -- 2) Não é possível vincular a um treino de outro atleta.
    UPDATE imported_activities SET workout_id = workout_b
    WHERE id = activity_a AND user_id = user_a
      AND EXISTS (
        SELECT 1 FROM workouts w
        JOIN training_plans tp ON tp.id = w.training_plan_id
        JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
        WHERE w.id = workout_b AND ap.user_id = user_a AND tp.status IN ('active', 'draft', 'completed'));
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 0 THEN
        RAISE EXCEPTION 'linking to another athlete''s workout must not update any row, got %', affected;
    END IF;
    IF (SELECT workout_id FROM imported_activities WHERE id = activity_a) <> workout_a THEN
        RAISE EXCEPTION 'a rejected link must leave the previous link untouched';
    END IF;

    -- 3) A atividade de outro atleta nunca é alterada, mesmo com um treino válido.
    UPDATE imported_activities SET workout_id = workout_a
    WHERE id = activity_b AND user_id = user_a
      AND EXISTS (
        SELECT 1 FROM workouts w
        JOIN training_plans tp ON tp.id = w.training_plan_id
        JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
        WHERE w.id = workout_a AND ap.user_id = user_a AND tp.status IN ('active', 'draft', 'completed'));
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 0 OR (SELECT workout_id FROM imported_activities WHERE id = activity_b) IS NOT NULL THEN
        RAISE EXCEPTION 'another user''s activity must not be linkable, got %', affected;
    END IF;

    -- 4) Um treino de plano cancelado não é aceito como alvo.
    UPDATE training_plans SET status = 'cancelled' WHERE id = plan_a;
    UPDATE imported_activities SET workout_id = NULL WHERE id = activity_a AND user_id = user_a;
    UPDATE imported_activities SET workout_id = workout_a
    WHERE id = activity_a AND user_id = user_a
      AND EXISTS (
        SELECT 1 FROM workouts w
        JOIN training_plans tp ON tp.id = w.training_plan_id
        JOIN athlete_profiles ap ON ap.id = tp.athlete_profile_id
        WHERE w.id = workout_a AND ap.user_id = user_a AND tp.status IN ('active', 'draft', 'completed'));
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 0 THEN
        RAISE EXCEPTION 'a workout of a cancelled plan must not be a link target, got %', affected;
    END IF;

    -- 5) Desvincular limpa o campo e a listagem volta a não ter treino.
    UPDATE training_plans SET status = 'active' WHERE id = plan_a;
    UPDATE imported_activities SET workout_id = workout_a WHERE id = activity_a;
    UPDATE imported_activities SET workout_id = NULL WHERE id = activity_a AND user_id = user_a;
    GET DIAGNOSTICS affected = ROW_COUNT;
    IF affected <> 1 OR (SELECT workout_id FROM imported_activities WHERE id = activity_a) IS NOT NULL THEN
        RAISE EXCEPTION 'unlinking should clear workout_id';
    END IF;
END;
$$;

ROLLBACK;
