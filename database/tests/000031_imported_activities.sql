BEGIN;

DO $$
DECLARE
    test_user_id UUID;
    other_user_id UUID;
    test_profile_id UUID;
    test_plan_id UUID;
    test_workout_id UUID;
    activity_id UUID;
    stored_hash TEXT;
    remaining_count INTEGER;
BEGIN
    INSERT INTO users (email, password_hash, display_name)
    VALUES ('imported-activities-test@example.invalid', 'test-only', 'Imported Activities Test')
    RETURNING id INTO test_user_id;

    INSERT INTO users (email, password_hash, display_name)
    VALUES ('imported-activities-other@example.invalid', 'test-only', 'Other User')
    RETURNING id INTO other_user_id;

    INSERT INTO athlete_profiles (user_id, experience_level)
    VALUES (test_user_id, 'intermediate')
    RETURNING id INTO test_profile_id;

    INSERT INTO training_plans (athlete_profile_id, starts_on, ends_on, status)
    VALUES (test_profile_id, DATE '2099-08-01', DATE '2099-08-28', 'active')
    RETURNING id INTO test_plan_id;

    INSERT INTO workouts (
        training_plan_id, scheduled_on, name, objective, duration_minutes,
        target_rpe, structure, explanation, status
    ) VALUES (
        test_plan_id, DATE '2099-08-10', 'Sweet spot', 'Teste', 60,
        6, '{}'::jsonb, '{"rules":[]}'::jsonb, 'planned'
    ) RETURNING id INTO test_workout_id;

    -- Uma atividade vinculada a um treino planejado é aceita e guarda o hash.
    INSERT INTO imported_activities (
        user_id, workout_id, source, file_hash, started_at, moving_seconds, distance_km
    ) VALUES (
        test_user_id, test_workout_id, 'fit', repeat('a', 64), now(), 3600, 35.5
    ) RETURNING id INTO activity_id;

    SELECT file_hash INTO stored_hash FROM imported_activities WHERE id = activity_id;
    IF stored_hash <> repeat('a', 64) THEN
        RAISE EXCEPTION 'file_hash was not stored: %', stored_hash;
    END IF;

    -- O mesmo arquivo (mesmo hash) não pode ser importado duas vezes pelo
    -- mesmo atleta.
    BEGIN
        INSERT INTO imported_activities (user_id, source, file_hash, started_at, moving_seconds)
        VALUES (test_user_id, 'fit', repeat('a', 64), now(), 1800);
        RAISE EXCEPTION 'duplicate file_hash for the same user should be rejected';
    EXCEPTION WHEN unique_violation THEN
        NULL;
    END;

    -- O mesmo hash já é aceito para outro atleta (arquivos de exemplo
    -- compartilhados, por exemplo, não devem travar dois usuários entre si).
    INSERT INTO imported_activities (user_id, source, file_hash, started_at, moving_seconds)
    VALUES (other_user_id, 'fit', repeat('a', 64), now(), 1800);

    -- Uma atividade sem treino correspondente é aceita, sem vínculo.
    INSERT INTO imported_activities (user_id, source, file_hash, started_at, moving_seconds)
    VALUES (test_user_id, 'gpx', repeat('b', 64), now(), 900);

    -- Apagar o treino planejado não apaga a atividade importada, só o vínculo.
    DELETE FROM workouts WHERE id = test_workout_id;
    IF (SELECT workout_id FROM imported_activities WHERE id = activity_id) IS NOT NULL THEN
        RAISE EXCEPTION 'workout_id should be cleared when the workout is deleted';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM imported_activities WHERE id = activity_id) THEN
        RAISE EXCEPTION 'the imported activity itself should survive workout deletion';
    END IF;

    -- Apagar o usuário apaga as atividades dele, em cascata.
    DELETE FROM users WHERE id = test_user_id;
    SELECT count(*) INTO remaining_count FROM imported_activities WHERE user_id = test_user_id;
    IF remaining_count <> 0 THEN
        RAISE EXCEPTION 'imported activities should be deleted in cascade with the user, found %', remaining_count;
    END IF;

    -- A atividade do outro usuário não foi afetada.
    IF NOT EXISTS (SELECT 1 FROM imported_activities WHERE user_id = other_user_id) THEN
        RAISE EXCEPTION 'the other user activity should not have been affected';
    END IF;
END;
$$;

ROLLBACK;
