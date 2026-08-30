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
    go run ./cmd/api
}
finally {
    Pop-Location
}
