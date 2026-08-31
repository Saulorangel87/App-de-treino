CREATE OR REPLACE FUNCTION complete_training_plan_when_terminal()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.status IN ('completed', 'skipped')
       AND OLD.status IS DISTINCT FROM NEW.status THEN
        UPDATE training_plans tp
        SET status = 'completed'
        WHERE tp.id = NEW.training_plan_id
          AND tp.status = 'active'
          AND NOT EXISTS (
              SELECT 1
              FROM workouts pending
              WHERE pending.training_plan_id = tp.id
                AND pending.status IN ('planned', 'adapted', 'in_progress')
          );
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER workouts_complete_training_plan
AFTER UPDATE OF status ON workouts
FOR EACH ROW
EXECUTE FUNCTION complete_training_plan_when_terminal();

UPDATE training_plans tp
SET status = 'completed'
WHERE tp.status = 'active'
  AND EXISTS (
      SELECT 1 FROM workouts w WHERE w.training_plan_id = tp.id
  )
  AND NOT EXISTS (
      SELECT 1
      FROM workouts pending
      WHERE pending.training_plan_id = tp.id
        AND pending.status IN ('planned', 'adapted', 'in_progress')
  );
