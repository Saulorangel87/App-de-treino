$ErrorActionPreference = 'Stop'

# Execute the actual repository SELECTs against synthetic CTEs in a read-only
# transaction. No real athlete data is read and no tables/fixtures are persisted.
$readinessRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$readinessSource = Get-Content -Raw -LiteralPath (Join-Path $readinessRoot 'backend/internal/repository/planning.go')
$readinessMatches = [regex]::Matches($readinessSource, 's\.pool\.QueryRow\(ctx,\s*`(?<sql>[^`]+)`, input\.ProfileID,\s*\)')
$readinessSessionMatches = @($readinessMatches | Where-Object { $_.Groups['sql'].Value.Contains('FROM workout_sessions ws') })
$readinessRecoveryMatches = @($readinessMatches | Where-Object { $_.Groups['sql'].Value.Contains('FROM recovery_data') })
if ($readinessSessionMatches.Count -ne 1 -or $readinessRecoveryMatches.Count -ne 1) {
    throw 'Não foi possível identificar as duas consultas de histórico. Atualize este teste junto do repositório.'
}
# Check both real SQL column counts and semantics (including nulls/zero values).
$readinessSessionQuery = $readinessSessionMatches[0].Groups['sql'].Value
$readinessRecoveryQuery = $readinessRecoveryMatches[0].Groups['sql'].Value
$readinessFixture = @'
WITH workout_sessions(id, athlete_profile_id, status, completed_at, duration_minutes, actual_rpe) AS (
    VALUES
    (1, 'full', 'completed', now(), 45, 4::numeric),
    (2, 'full', 'completed', now() - interval '1 day', 60, 6),
    (3, 'mixed', 'completed', now(), 45, 5),
    (4, 'mixed', 'completed', now(), 0, 0),
    (5, 'mixed', 'completed', now(), 30, NULL),
    (6, 'mixed', 'completed', now(), NULL, 8),
    (7, 'mixed', 'cancelled', now(), 60, 8),
    (8, 'mixed', 'completed', now() - interval '40 days', 60, 8),
    (9, 'other', 'completed', now(), 120, 10)
), feedback(workout_session_id, fatigue_after, pain_reported) AS (
    VALUES (1, 2, false), (2, 2, false), (3, 2, false),
    (6, 5, true), (7, 5, true), (8, 5, true), (9, 5, true)
), recovery_data(athlete_profile_id, recorded_on, fatigue_level) AS (
    VALUES
    ('full', CURRENT_DATE, 2), ('full', CURRENT_DATE - 1, 2),
    ('mixed', CURRENT_DATE, NULL), ('mixed', CURRENT_DATE - 1, 4),
    ('mixed', CURRENT_DATE - 28, 5), ('checkins_only', CURRENT_DATE, 3),
    ('other', CURRENT_DATE, 5)
)
'@
$readinessCases = @(
    @{ Name = 'full'; Sessions = "(2, 105, 5::double precision, 2::double precision, false, 2, 2, 2, 2)"; Recovery = '(2, 2::double precision, 2)' },
    @{ Name = 'mixed'; Sessions = "(4, 75, (13.0/3)::double precision, 3.5::double precision, true, 2, 2, 2, 1)"; Recovery = '(2, 4::double precision, 1)' },
    @{ Name = 'empty'; Sessions = '(0, 0, 0::double precision, 0::double precision, false, 0, 0, 0, 0)'; Recovery = '(0, 0::double precision, 0)' },
    @{ Name = 'checkins_only'; Sessions = '(0, 0, 0::double precision, 0::double precision, false, 0, 0, 0, 0)'; Recovery = '(1, 3::double precision, 1)' }
)
$readinessSQL = "BEGIN READ ONLY;`n"
foreach ($readinessCase in $readinessCases) {
    # Names are fixed synthetic identifiers above, never user-supplied values.
    $readinessName = $readinessCase.Name
    $readinessSessionSQL = $readinessSessionQuery.Replace('$1', "'$readinessName'")
    $readinessRecoverySQL = $readinessRecoveryQuery.Replace('$1', "'$readinessName'")
    $readinessSQL += @"
DO `$readiness`$
DECLARE
    sessions record;
    recovery record;
BEGIN
    SELECT * INTO sessions FROM (
        $readinessFixture
        $readinessSessionSQL
    ) AS result(completed_sessions, completed_minutes, average_rpe, average_fatigue, pain_reported,
        sessions_with_duration, sessions_with_rpe, sessions_with_feedback, complete_sessions);
    IF ROW(sessions.completed_sessions, sessions.completed_minutes, sessions.average_rpe,
        sessions.average_fatigue, sessions.pain_reported, sessions.sessions_with_duration,
        sessions.sessions_with_rpe, sessions.sessions_with_feedback, sessions.complete_sessions)
        IS DISTINCT FROM ROW$($readinessCase.Sessions) THEN
        RAISE EXCEPTION 'readiness session fixture $readinessName failed: %', sessions;
    END IF;
    SELECT * INTO recovery FROM (
        $readinessFixture
        $readinessRecoverySQL
    ) AS result(recovery_checkins, average_recovery_fatigue, recovery_with_fatigue);
    IF ROW(recovery.recovery_checkins, recovery.average_recovery_fatigue, recovery.recovery_with_fatigue)
        IS DISTINCT FROM ROW$($readinessCase.Recovery) THEN
        RAISE EXCEPTION 'readiness recovery fixture $readinessName failed: %', recovery;
    END IF;
    RAISE NOTICE 'readiness fixture ${readinessName}: OK';
END
`$readiness`$;
"@
}
$readinessSQL += "`nROLLBACK;"

Push-Location $readinessRoot
try {
    $readinessSQL | docker compose exec -T postgres sh -c 'psql -X -q -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1'
    if ($LASTEXITCODE -ne 0) {
        throw 'Falha na validação das consultas de prontidão no PostgreSQL local.'
    }
}
finally {
    Pop-Location
}
