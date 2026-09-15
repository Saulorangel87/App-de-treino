$ErrorActionPreference = 'Stop'

# Execute the Evolution repository queries against synthetic CTEs in a
# read-only transaction. No real athlete data is read or persisted.
$evolutionRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$evolutionSource = Get-Content -Raw -LiteralPath (Join-Path $evolutionRoot 'backend/internal/repository/evolution.go')

function Get-EvolutionQuery([string]$pattern, [string]$label) {
    $match = [regex]::Match($evolutionSource, $pattern)
    if (-not $match.Success) {
        throw "Consulta de Evolução '$label' não encontrada. Atualize este teste junto do repositório."
    }
    return $match.Groups['sql'].Value.Trim().Replace('$1', "'user-1'")
}

$summaryQuery = Get-EvolutionQuery 's\.pool\.QueryRow\(ctx,\s*`(?<sql>[^`]+)`, userID,\s*\)' 'summary'
$weeksQuery = Get-EvolutionQuery 'weeks, err := s\.pool\.Query\(ctx,\s*`(?<sql>[^`]+)`, userID\)' 'weeks'
$goalProgressQuery = Get-EvolutionQuery 'goalRows, err := s\.pool\.Query\(ctx,\s*`(?<sql>[^`]+)`, userID\)' 'goal progress'
$recentQuery = Get-EvolutionQuery 'recentRows, err := s\.pool\.Query\(ctx,\s*`(?<sql>[^`]+)`, userID\)' 'recent sessions'
$recoveryQuery = Get-EvolutionQuery 'recoveryRows, err := s\.pool\.Query\(ctx,\s*`(?<sql>[^`]+)`, userID\)' 'recovery'

$evolutionFixtures = @'
WITH athlete_profiles(id, user_id) AS (
    VALUES ('profile-1', 'user-1'), ('profile-2', 'other-user')
), workouts(id, name, duration_minutes, target_rpe, explanation) AS (
    VALUES
    ('workout-past', 'Endurance passado', 40, 5::numeric, '{}'::jsonb),
    ('workout-future', 'Endurance futuro', 120, 10::numeric, '{}'::jsonb),
    ('workout-cancelled', 'Cancelado passado', 30, 4::numeric, '{}'::jsonb),
    ('workout-future-cancelled', 'Cancelado futuro', 30, 4::numeric, '{}'::jsonb),
    ('workout-other', 'Outro atleta', 90, 6::numeric, '{}'::jsonb)
), workout_sessions(
    id, workout_id, athlete_profile_id, status, completed_at, cancelled_at,
    duration_minutes, actual_rpe, distance_km, elevation_gain_m,
    average_power_watts, average_heart_rate, created_at
) AS (
    VALUES
    ('session-past', 'workout-past', 'profile-1', 'completed', date_trunc('week', now()), NULL, 40, 5, 20::numeric, 100, 200, 140, date_trunc('week', now())),
    ('session-future', 'workout-future', 'profile-1', 'completed', now() + interval '2 hours', NULL, 120, 10, 80::numeric, 500, 400, 180, now()),
    ('session-cancelled', 'workout-cancelled', 'profile-1', 'cancelled', NULL, date_trunc('week', now()), NULL, NULL, NULL, NULL, NULL, NULL, date_trunc('week', now())),
    ('session-future-cancelled', 'workout-future-cancelled', 'profile-1', 'cancelled', NULL, now() + interval '2 hours', NULL, NULL, NULL, NULL, NULL, NULL, now()),
    ('session-other', 'workout-other', 'profile-2', 'completed', now() - interval '1 day', NULL, 90, 6, 45::numeric, 200, 250, 145, now() - interval '1 day')
), feedback(workout_session_id, fatigue_after, pain_reported) AS (
    VALUES ('session-past', 2, false), ('session-future', 5, true)
), goals(athlete_profile_id, goal_type, priority, target_date, details) AS (
    VALUES
    ('profile-1', 'endurance', 1, NULL::date, '{"notes":"Meta de resistência"}'::jsonb),
    ('profile-1', 'performance', 2, NULL::date, '{"notes":"Meta secundária"}'::jsonb)
), recovery_data(athlete_profile_id, recorded_on, sleep_minutes, sleep_quality, stress_level, fatigue_level) AS (
    VALUES
    ('profile-1', CURRENT_DATE, 480, 4, 2, 2),
    ('profile-1', CURRENT_DATE - 1, 300, 3, 2, 2),
    ('profile-1', CURRENT_DATE + 1, 240, 1, 5, 5),
    ('other-profile', CURRENT_DATE, 240, 1, 5, 5)
)
'@

function Compose-EvolutionQuery([string]$query) {
    if ($query.StartsWith('WITH ')) {
        return "$evolutionFixtures,`n$($query.Substring(5))"
    }
    return "$evolutionFixtures`n$query"
}

function Invoke-EvolutionQuery([string]$query) {
    $sql = "BEGIN READ ONLY;`n$(Compose-EvolutionQuery $query);`nROLLBACK;"
    Push-Location $evolutionRoot
    try {
        $output = $sql | docker compose exec -T postgres sh -c 'psql -X -q -A -t -F "|" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1'
        if ($LASTEXITCODE -ne 0) {
            throw 'Falha ao executar uma consulta de Evolução no PostgreSQL local.'
        }
        return @($output | ForEach-Object { $_.Trim() } | Where-Object { $_ })
    }
    finally {
        Pop-Location
    }
}

$summaryActual = @(Invoke-EvolutionQuery $summaryQuery)
if ($summaryActual.Count -ne 1 -or $summaryActual[0] -cne '1|1|40|5|2|20|100|200|140|200|30') {
    throw "Resumo de Evolução divergente: $($summaryActual -join '; ')"
}
Write-Output 'evolution summary excludes future and other-athlete sessions: OK'

$currentWeekQuery = "SELECT * FROM ($weeksQuery) AS weekly WHERE weekly.week_start = date_trunc('week', CURRENT_DATE)::date::text"
$weekActual = @(Invoke-EvolutionQuery $currentWeekQuery)
$weekColumns = if ($weekActual.Count -eq 1) { $weekActual[0].Split('|') } else { @() }
if ($weekColumns.Count -ne 11 -or $weekColumns[1..10] -join '|' -cne '1|1|40|5|20|100|200|140|200|30') {
    throw "Semana atual de Evolução divergente: $($weekActual -join '; ')"
}
Write-Output 'evolution weekly aggregate excludes future sessions: OK'

$goalProgressActual = @(Invoke-EvolutionQuery $goalProgressQuery)
if ($goalProgressActual.Count -ne 2 -or $goalProgressActual[0] -cne 'endurance|1||Meta de resistência' -or $goalProgressActual[1] -cne 'performance|2||Meta secundária') {
    throw "Objetivos de Evolução divergentes: $($goalProgressActual -join '; ')"
}
Write-Output 'evolution goal progress reads both current goals: OK'

$recentActual = @(Invoke-EvolutionQuery $recentQuery)
$recentColumns = if ($recentActual.Count -eq 1) { $recentActual[0].Split('|') } else { @() }
if ($recentColumns.Count -ne 11 -or $recentColumns[1] -cne 'Endurance passado' -or
    $recentColumns[2..9] -join '|' -cne '40|5|40|5|20|200|140|2') {
    throw "Sessões recentes de Evolução divergentes: $($recentActual -join '; ')"
}
Write-Output 'evolution recent sessions exclude future sessions: OK'

$recoveryActual = @(Invoke-EvolutionQuery $recoveryQuery)
if ($recoveryActual.Count -ne 2 -or $recoveryActual[0] -notmatch '^\d{4}-\d{2}-\d{2}\|480\|4\|2\|2\|ready$' -or
    $recoveryActual[1] -notmatch '^\d{4}-\d{2}-\d{2}\|300\|3\|2\|2\|caution$') {
    throw "Pontos de recuperação de Evolução divergentes: $($recoveryActual -join '; ')"
}
Write-Output 'evolution recovery points exclude future check-ins: OK'
