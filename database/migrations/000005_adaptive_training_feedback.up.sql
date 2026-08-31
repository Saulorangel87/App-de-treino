CREATE OR REPLACE FUNCTION adapt_training_plan_after_feedback()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    source_workout_id UUID;
    source_plan_id UUID;
    source_profile_id UUID;
    source_date DATE;
    source_target_rpe NUMERIC(3,1);
    actual_rpe NUMERIC(3,1);
    adaptation_kind TEXT;
    affected_sessions INTEGER := 0;
    duration_factor NUMERIC := 1;
    target_rpe_delta NUMERIC := 0;
    target_rpe_cap NUMERIC;
    minimum_target_rpe NUMERIC := 0;
    adaptation_reason TEXT;
    safety_notice TEXT;
    candidate RECORD;
    new_duration INTEGER;
    new_target_rpe NUMERIC(3,1);
    new_explanation JSONB;
BEGIN
    SELECT ws.workout_id, ws.athlete_profile_id, ws.actual_rpe,
           w.training_plan_id, w.scheduled_on, w.target_rpe
    INTO source_workout_id, source_profile_id, actual_rpe,
         source_plan_id, source_date, source_target_rpe
    FROM workout_sessions ws
    JOIN workouts w ON w.id = ws.workout_id
    WHERE ws.id = NEW.workout_session_id;

    IF source_workout_id IS NULL THEN
        RETURN NEW;
    END IF;

    IF NEW.pain_reported THEN
        adaptation_kind := 'safety';
        affected_sessions := 2;
        duration_factor := 0.80;
        target_rpe_cap := 3;
        minimum_target_rpe := 2;
        adaptation_reason := 'Carga reduzida porque houve relato de dor após a sessão anterior.';
        safety_notice := 'Interrompa o treino se a dor reaparecer. Dor persistente ou intensa deve ser avaliada por um profissional de saúde.';
    ELSIF actual_rpe >= 9 OR NEW.fatigue_after = 5 OR NEW.difficulty = 'very_hard' THEN
        adaptation_kind := 'recovery';
        affected_sessions := 2;
        duration_factor := 0.80;
        target_rpe_delta := -1;
        target_rpe_cap := 4;
        minimum_target_rpe := 3;
        adaptation_reason := 'Carga reduzida porque a sessão anterior gerou esforço ou fadiga muito altos.';
    ELSIF actual_rpe >= source_target_rpe + 2 OR NEW.fatigue_after >= 4 OR NEW.difficulty = 'hard' THEN
        adaptation_kind := 'recovery';
        affected_sessions := 1;
        duration_factor := 0.90;
        target_rpe_delta := -1;
        minimum_target_rpe := 3;
        adaptation_reason := 'A próxima carga foi reduzida porque o esforço percebido ficou acima do esperado.';
    ELSIF actual_rpe <= source_target_rpe - 2
          AND NEW.fatigue_after <= 2
          AND NEW.difficulty IN ('easy', 'very_easy') THEN
        adaptation_kind := 'progression';
        affected_sessions := 1;
        duration_factor := 1.05;
        adaptation_reason := 'A próxima sessão recebeu uma progressão leve porque o treino anterior foi claramente mais fácil que o previsto.';
    END IF;

    IF affected_sessions = 0 THEN
        RETURN NEW;
    END IF;

    FOR candidate IN
        SELECT w.id, w.duration_minutes, w.target_rpe, w.explanation,
               a.available_minutes
        FROM workouts w
        LEFT JOIN availability a
          ON a.athlete_profile_id = source_profile_id
         AND a.weekday = EXTRACT(DOW FROM w.scheduled_on)::smallint
        WHERE w.training_plan_id = source_plan_id
          AND w.scheduled_on > source_date
          AND w.status = 'planned'
        ORDER BY w.scheduled_on, w.created_at
        LIMIT affected_sessions
        FOR UPDATE OF w
    LOOP
        new_duration := GREATEST(20, ROUND(candidate.duration_minutes * duration_factor)::integer);
        IF candidate.available_minutes IS NOT NULL AND candidate.available_minutes > 0 THEN
            new_duration := LEAST(new_duration, candidate.available_minutes);
        END IF;

        new_target_rpe := candidate.target_rpe + target_rpe_delta;
        IF target_rpe_cap IS NOT NULL THEN
            new_target_rpe := LEAST(new_target_rpe, target_rpe_cap);
        END IF;
        new_target_rpe := GREATEST(new_target_rpe, minimum_target_rpe);

        new_explanation := jsonb_set(
            COALESCE(candidate.explanation, '{}'::jsonb),
            '{adaptation}',
            jsonb_strip_nulls(jsonb_build_object(
                'kind', adaptation_kind,
                'reason', adaptation_reason,
                'safety_notice', safety_notice,
                'source_workout_id', source_workout_id,
                'previous_duration_minutes', candidate.duration_minutes,
                'previous_target_rpe', candidate.target_rpe
            )),
            true
        );
        new_explanation := jsonb_set(
            new_explanation,
            '{rules}',
            COALESCE(new_explanation->'rules', '[]'::jsonb) || jsonb_build_array(adaptation_reason),
            true
        );

        UPDATE workouts
        SET duration_minutes = new_duration,
            target_rpe = new_target_rpe,
            explanation = new_explanation
        WHERE id = candidate.id;
    END LOOP;

    RETURN NEW;
END;
$$;

CREATE TRIGGER feedback_adapts_future_workouts
AFTER INSERT ON feedback
FOR EACH ROW
EXECUTE FUNCTION adapt_training_plan_after_feedback();
