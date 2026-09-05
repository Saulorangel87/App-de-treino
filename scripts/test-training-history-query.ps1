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
workouts(id, training_plan_id, scheduled_on, status) AS (
    VALUES
        (1, 1, CURRENT_DATE - 1, 'completed'),
        (2, 1, CURRENT_DATE - 2, 'skipped'),
        (3, 1, CURRENT_DATE - 3, 'planned'),
        (4, 1, CURRENT_DATE - 4, 'adapted'),
        (5, 1, CURRENT_DATE - 5, 'in_progress'),
        (6, 1, CURRENT_DATE, 'planned'),
        (7, 1, CURRENT_DATE, 'completed'),
        (8, 1, CURRENT_DATE, 'skipped'),
        (9, 1, CURRENT_DATE - 8, 'completed'),
        (10, 2, CURRENT_DATE - 20, 'completed'),
        (11, 2, CURRENT_DATE - 35, 'completed'),
        (12, 3, CURRENT_DATE - 1, 'completed'),
        (13, 4, CURRENT_DATE - 1, 'completed'),
        (14, 5, CURRENT_DATE - 1, 'completed')
),
workout_sessions(id, athlete_profile_id, status, completed_at, duration_minutes, actual_rpe) AS (
    VALUES
        (1, 'profile-test', 'completed', now() - interval '1 day', 45, 4::numeric),
        (2, 'profile-test', 'completed', now() - interval '2 days', 60, 5),
        (3, 'profile-test', 'completed', now() - interval '5 days', 0, 6),
        (4, 'profile-test', 'completed', now() - interval '8 days', 90, 6),
        (5, 'profile-test', 'completed', now() - interval '20 days', NULL, 5),
        (6, 'profile-test', 'completed', now() - interval '35 days', 30, NULL),
        (7, 'profile-test', 'completed', now() + interval '1 day', 120, 8),
        (8, 'other-profile', 'completed', now() - interval '1 day', 120, 8),
        (9, 'profile-test', 'cancelled', NULL, 120, 8)
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
$historyExpected = @(
    '7|7|2|2|2|1|3|105|2|1|480',
    '28|9|4|2|2|1|5|195|3|2|1020',
    '42|10|5|2|2|1|6|225|3|3|1020'
)
if ($historyActual.Count -ne $historyExpected.Count) {
    throw "Quantidade inesperada de janelas: $($historyActual -join '; ')"
}
for ($historyIndex = 0; $historyIndex -lt $historyExpected.Count; $historyIndex++) {
    if ($historyActual[$historyIndex] -cne $historyExpected[$historyIndex]) {
        throw "Janela divergente. Esperado '$($historyExpected[$historyIndex])'; recebido '$($historyActual[$historyIndex])'."
    }
}
Write-Output 'training history fixtures 7d/28d/42d: OK'
