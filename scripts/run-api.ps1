$ErrorActionPreference = 'Stop'

$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$environmentFile = Join-Path $projectRoot '.env'

if (-not (Test-Path -LiteralPath $environmentFile)) {
    throw 'Arquivo .env não encontrado. Copie .env.example para .env antes de iniciar a API.'
}

Get-Content -LiteralPath $environmentFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -and -not $line.StartsWith('#')) {
        $parts = $line.Split('=', 2)
        if ($parts.Count -eq 2) {
            Set-Item -LiteralPath "Env:$($parts[0])" -Value $parts[1]
        }
    }
}

Push-Location (Join-Path $projectRoot 'backend')
try {
    # O Smart App Control pode bloquear o executável temporário criado pelo
    # `go run`; o .gotmp é ignorado pelo Git e fica dentro do workspace.
    $temporaryRoot = Join-Path (Get-Location) '.gotmp'
    $binaryPath = Join-Path $temporaryRoot 'cadencia-api.exe'
    $goCachePath = Join-Path $temporaryRoot 'go-cache'
    New-Item -ItemType Directory -Force -Path $temporaryRoot, $goCachePath | Out-Null

    $previousGoCache = $env:GOCACHE
    $env:GOCACHE = $goCachePath
    try {
        go build -o $binaryPath ./cmd/api
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
        & $binaryPath
    }
    finally {
        $env:GOCACHE = $previousGoCache
    }
}
finally {
    Pop-Location
}
