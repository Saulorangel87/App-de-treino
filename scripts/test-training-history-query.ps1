$ErrorActionPreference = 'Stop'

# Run the exact repository query against synthetic CTEs in a read-only
# transaction. The CTE names shadow the tables, so no athlete record is read.
$historyRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$historySource = Get-Content -Raw -LiteralPath (Join-Path $historyRoot 'backend/internal/repository/planning.go')
$historyMatch = [regex]::Match($historySource, 'const planningTrainingHistoryQuery = `(?<sql>[^`]+)`')
if (-not $historyMatch.Success) {
    throw 'Consulta planningTrainingHistoryQuery não encontrada. Atualize este teste junto do repositório.'
}
$historyQuery = $historyMatch.Groups['sql'].Value.Trim()
if (-not $historyQuery.StartsWith('WITH window_sizes')) {
    throw 'Formato inesperado da consulta de histórico. Atualize o injetor de fixtures.'
}
$historyQuery = $historyQuery.Replace('$1', "'profile-test'")

$historyFixtures = @'
WITH training_plans(id, athlete_profile_id, status) AS (
    VALUES
        (1, 'profile-test', 'active'),
        (2, 'profile-test', 'completed'),
        (3, 'profile-test', 'draft'),
        (4, 'other-profile', 'active'),
        (5, 'profile-test', 'cancelled')
),
workouts(id, training_plan_id, scheduled_on, status, target_rpe) AS (
    VALUES
        (1, 1, CURRENT_DATE - 1, 'completed', 4::numeric),
        (2, 1, CURRENT_DATE - 2, 'skipped', 5),
        (3, 1, CURRENT_DATE - 3, 'planned', 5),
        (4, 1, CURRENT_DATE - 4, 'adapted', 5),
        (5, 1, CURRENT_DATE - 5, 'in_progress', 6),
        (6, 1, CURRENT_DATE, 'planned', 5),
        (7, 1, CURRENT_DATE, 'completed', 6),
        (8, 1, CURRENT_DATE, 'skipped', 5),
        (9, 1, CURRENT_DATE - 8, 'completed', 6),
        (10, 2, CURRENT_DATE - 20, 'completed', 5),
        (11, 2, CURRENT_DATE - 35, 'completed', 5),
        (12, 3, CURRENT_DATE - 1, 'completed', 5),
        (13, 4, CURRENT_DATE - 1, 'completed', 5),
        (14, 5, CURRENT_DATE - 1, 'completed', 5)
),
workout_sessions(id, workout_id, athlete_profile_id, status, completed_at, duration_minutes, actual_rpe) AS (
    VALUES
        (1, 1, 'profile-test', 'completed', now() - interval '1 day', 45, 4::numeric),
        (2, 2, 'profile-test', 'completed', now() - interval '2 days', 60, 7),
        (3, 5, 'profile-test', 'completed', now() - interval '5 days', 0, 6),
        (4, 9, 'profile-test', 'completed', now() - interval '8 days', 90, 6),
        (5, 10, 'profile-test', 'completed', now() - interval '20 days', NULL, 5),
        (6, 11, 'profile-test', 'completed', now() - interval '35 days', 30, NULL),
        (7, 7, 'profile-test', 'completed', now() + interval '1 day', 120, 8),
        (8, 13, 'other-profile', 'completed', now() - interval '1 day', 120, 8),
        (9, 1, 'profile-test', 'cancelled', NULL, 120, 8)
),
feedback(id, workout_session_id, pain_reported, fatigue_after) AS (
    VALUES
        (1, 1, false, 2),
        (2, 2, true, 4),
        (3, 3, false, 5),
        (4, 4, false, 3),
        (5, 7, true, 5),
        (6, 5, true, NULL)
),
recovery_data(id, athlete_profile_id, recorded_on, sleep_minutes, sleep_quality, stress_level, fatigue_level) AS (
    VALUES
        (1, 'profile-test', CURRENT_DATE, 480, 4, 2, 2),
        (2, 'profile-test', CURRENT_DATE - 2, 300, 3, 2, 2),
        (3, 'profile-test', CURRENT_DATE - 4, 420, 2, 4, 4),
        (4, 'profile-test', CURRENT_DATE - 10, 450, 4, 2, 2),
        (5, 'profile-test', CURRENT_DATE - 30, 450, 4, 2, 5),
        (6, 'profile-test', CURRENT_DATE + 1, 480, 4, 2, 2),
        (7, 'profile-test', CURRENT_DATE - 3, NULL, 4, 2, 2),
        (8, 'other-profile', CURRENT_DATE, 240, 1, 5, 5)
),
'@
$historySyntheticQuery = $historyQuery -replace '^WITH ', $historyFixtures
$historySQL = "BEGIN READ ONLY;`n$historySyntheticQuery;`nROLLBACK;"

Push-Location $historyRoot
try {
    $historyOutput = $historySQL | docker compose exec -T postgres sh -c 'psql -X -q -A -t -F "|" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1'
    if ($LASTEXITCODE -ne 0) {
        throw 'Falha ao executar a consulta de histórico no PostgreSQL local.'
    }
}
finally {
    Pop-Location
}

$historyActual = @($historyOutput | ForEach-Object { $_.Trim() } | Where-Object { $_ })
$historyExpectedPrefix = @(
    '7|7|2|2|2|1|3|105|2|1|600|3|3|1|2|1|4|3|2|1',
    '28|9|4|2|2|1|5|195|3|2|1140|5|4|2|2|1|5|4|2|1',
    '42|10|5|2|2|1|6|225|3|3|1140|5|4|2|2|1|6|5|3|2'
)
if ($historyActual.Count -ne $historyExpectedPrefix.Count) {
    throw "Quantidade inesperada de janelas: $($historyActual -join '; ')"
}
for ($historyIndex = 0; $historyIndex -lt $historyExpectedPrefix.Count; $historyIndex++) {
    $historyColumns = $historyActual[$historyIndex].Split('|')
    $historyPrefix = ($historyColumns[0..19] -join '|')
    if ($historyPrefix -cne $historyExpectedPrefix[$historyIndex]) {
        throw "Janela divergente. Esperado prefixo '$($historyExpectedPrefix[$historyIndex])'; recebido '$($historyActual[$historyIndex])'."
    }
    if ($historyColumns.Count -ne 28 -or $historyColumns[20] -notmatch 'Z|\+00$' -or
        $historyColumns[21] -cne '1' -or $historyColumns[22] -notmatch 'Z|\+00$' -or
        $historyColumns[23] -cne '1' -or $historyColumns[24] -notmatch '^\d{4}-\d{2}-\d{2}$' -or
        $historyColumns[25] -cne '0' -or $historyColumns[26] -cne '1' -or $historyColumns[27] -cne '1') {
        throw "Metadados temporais divergentes: '$($historyActual[$historyIndex])'."
    }
}
Write-Output 'training history v2 fixtures 7d/28d/42d and temporal quality: OK'
