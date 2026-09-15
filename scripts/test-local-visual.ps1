$ErrorActionPreference = 'Stop'

# Smoke test visual local. Cria uma conta descartável, preenche o onboarding,
# abre as rotas principais em Chrome headless e remove a conta ao terminar.
$projectRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$frontendURL = 'http://localhost:3000'
$apiURL = 'http://localhost:8080'
$origin = $frontendURL
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$password = 'visual-test-000030'
$email = "visual-integration-$([guid]::NewGuid().ToString('N').Substring(0, 8))@example.invalid"
$registered = $false
$chromeProcess = $null
$webSocket = $null
$temporaryRoot = Join-Path ([IO.Path]::GetTempPath()) ('cadencia-visual-' + [guid]::NewGuid().ToString('N'))
$chromeProfile = Join-Path $temporaryRoot 'chrome-profile'
$screenshots = Join-Path $temporaryRoot 'screenshots'
$cdpPort = Get-Random -Minimum 9223 -Maximum 9399
$script:cdpID = 0

New-Item -ItemType Directory -Force -Path $screenshots | Out-Null

function Invoke-API {
    param(
        [Parameter(Mandatory)] [string]$Method,
        [Parameter(Mandatory)] [string]$Path,
        [object]$Body,
        [int[]]$ExpectedStatus = @(200)
    )

    $request = @{
        UseBasicParsing = $true
        SkipHttpErrorCheck = $true
        Uri = "$apiURL$Path"
        Method = $Method
        WebSession = $session
        Headers = @{ Origin = $origin }
        ErrorAction = 'Stop'
    }
    if ($null -ne $Body) {
        $request.ContentType = 'application/json'
        $request.Body = $Body | ConvertTo-Json -Depth 30 -Compress
    }
    $response = Invoke-WebRequest @request
    if ($ExpectedStatus -notcontains [int]$response.StatusCode) {
        throw "$Method $Path retornou $($response.StatusCode), esperado $($ExpectedStatus -join ', '). Corpo: $($response.Content)"
    }
    if ([string]::IsNullOrWhiteSpace($response.Content)) {
        return $null
    }
    return $response.Content | ConvertFrom-Json
}

function Receive-CDPMessage {
    $buffer = New-Object byte[] 65536
    $builder = [Text.StringBuilder]::new()
    do {
        $segment = [ArraySegment[byte]]::new($buffer)
        $received = $webSocket.ReceiveAsync($segment, [Threading.CancellationToken]::None).GetAwaiter().GetResult()
        if ($received.MessageType -eq [Net.WebSockets.WebSocketMessageType]::Close) {
            throw 'Chrome encerrou a sessão CDP inesperadamente.'
        }
        [void]$builder.Append([Text.Encoding]::UTF8.GetString($buffer, 0, $received.Count))
    } while (-not $received.EndOfMessage)
    return $builder.ToString() | ConvertFrom-Json
}

function Invoke-CDP {
    param(
        [Parameter(Mandatory)] [string]$Method,
        [hashtable]$Params = @{}
    )
    $script:cdpID++
    $id = $script:cdpID
    $payload = @{ id = $id; method = $Method; params = $Params } | ConvertTo-Json -Depth 30 -Compress
    $bytes = [Text.Encoding]::UTF8.GetBytes($payload)
    $segment = [ArraySegment[byte]]::new($bytes)
    [void]$webSocket.SendAsync($segment, [Net.WebSockets.WebSocketMessageType]::Text, $true, [Threading.CancellationToken]::None).GetAwaiter().GetResult()
    do {
        $message = Receive-CDPMessage
    } while ($message.id -ne $id)
    if ($message.error) {
        throw "CDP $Method falhou: $($message.error.message)"
    }
    return $message
}

function Get-PageText {
    $result = Invoke-CDP -Method 'Runtime.evaluate' -Params @{
        expression = 'document.body ? document.body.innerText : ""'
        returnByValue = $true
    }
    return [string]$result.result.result.value
}

function Dismiss-NewsModal {
    $result = Invoke-CDP -Method 'Runtime.evaluate' -Params @{
        expression = "(() => { const button = Array.from(document.querySelectorAll('button')).find((item) => item.textContent?.trim() === 'Entendi, continuar'); if (!button) return false; button.click(); return true; })()"
        returnByValue = $true
    }
    if ($result.result.result.value) {
        Start-Sleep -Milliseconds 150
    }
}

function Assert-DOMSelector {
    param([Parameter(Mandatory)] [string]$Selector)
    $result = Invoke-CDP -Method 'Runtime.evaluate' -Params @{
        expression = "Boolean(document.querySelector('$Selector'))"
        returnByValue = $true
    }
    if (-not $result.result.result.value) {
        throw "Seletor esperado não encontrado: $Selector"
    }
}

function Click-DOMSelector {
    param([Parameter(Mandatory)] [string]$Selector)
    $result = Invoke-CDP -Method 'Runtime.evaluate' -Params @{
        expression = "(() => { const element = document.querySelector('$Selector'); if (!element) return false; element.click(); return true; })()"
        returnByValue = $true
    }
    if (-not $result.result.result.value) {
        throw "Elemento esperado não encontrado para clique: $Selector"
    }
    Start-Sleep -Milliseconds 150
}

function Scroll-DOMSelector {
    param([Parameter(Mandatory)] [string]$Selector)
    $result = Invoke-CDP -Method 'Runtime.evaluate' -Params @{
        expression = "(() => { const element = document.querySelector('$Selector'); if (!element) return false; element.scrollIntoView({ block: 'start' }); return true; })()"
        returnByValue = $true
    }
    if (-not $result.result.result.value) {
        throw "Elemento esperado não encontrado para rolagem: $Selector"
    }
    Start-Sleep -Milliseconds 150
}

function Wait-PageText {
    param([Parameter(Mandatory)] [string]$Text)
    for ($attempt = 0; $attempt -lt 40; $attempt++) {
        $body = Get-PageText
        if ($body.Contains($Text)) {
            return $body
        }
        Start-Sleep -Milliseconds 250
    }
    $preview = $body -replace '\s+', ' '
    if ($preview.Length -gt 500) { $preview = $preview.Substring(0, 500) }
    throw "Texto não encontrado na página: $Text. Conteúdo: $preview"
}

function Open-Page {
    param(
        [Parameter(Mandatory)] [string]$Path,
        [Parameter(Mandatory)] [string]$ReadyText,
        [Parameter(Mandatory)] [int]$Width,
        [Parameter(Mandatory)] [int]$Height,
        [Parameter(Mandatory)] [string]$ScreenshotPath
    )
    Invoke-CDP -Method 'Emulation.setDeviceMetricsOverride' -Params @{
        width = $Width
        height = $Height
        deviceScaleFactor = 1
        mobile = ($Width -lt 600)
    } | Out-Null
    Invoke-CDP -Method 'Page.navigate' -Params @{ url = "$frontendURL$Path" } | Out-Null
    $body = Wait-PageText -Text $ReadyText
    Dismiss-NewsModal
    $body = Get-PageText
    $image = Invoke-CDP -Method 'Page.captureScreenshot' -Params @{ format = 'png' }
    [IO.File]::WriteAllBytes($ScreenshotPath, [Convert]::FromBase64String($image.result.data))
    return $body
}

function Assert-Contains {
    param([string]$Value, [string]$Expected)
    if (-not $Value.Contains($Expected)) {
        throw "Texto esperado não encontrado: $Expected"
    }
}

try {
    $registration = Invoke-API -Method 'POST' -Path '/v1/auth/register' -Body @{
        email = $email
        password = $password
        display_name = 'Visual Integration'
    } -ExpectedStatus @(201, 503)
    $registered = $true
    Write-Output "temporary visual account created: $email"

    Invoke-API -Method 'PUT' -Path '/v1/profile' -Body @{
        birth_date = $null
        sex = $null
        height_cm = 178
        weight_kg = 76
        waist_cm = 82
        body_fat_percent = 18
        weight_trend = 'stable'
        experience_level = 'beginner'
        activity_level = 'regular'
    } | Out-Null
    Invoke-API -Method 'PUT' -Path '/v1/onboarding/limitations' -Body @{ limitations = @() } | Out-Null
    Invoke-API -Method 'PUT' -Path '/v1/onboarding/goals' -Body @{ goals = @(
            @{ goal_type = 'health'; priority = 1; target_date = $null; details = 'Consistência' },
            @{ goal_type = 'performance'; priority = 2; target_date = $null; details = 'Evoluir com segurança' }
        ) } | Out-Null
    Invoke-API -Method 'PUT' -Path '/v1/onboarding/availability' -Body @{ availability = @(
            @{ weekday = 0; available_minutes = 0; preferred_time = $null; location = $null },
            @{ weekday = 1; available_minutes = 60; preferred_time = '06:30'; location = 'indoor' },
            @{ weekday = 2; available_minutes = 0; preferred_time = $null; location = $null },
            @{ weekday = 3; available_minutes = 0; preferred_time = $null; location = $null },
            @{ weekday = 4; available_minutes = 0; preferred_time = $null; location = $null },
            @{ weekday = 5; available_minutes = 0; preferred_time = $null; location = $null },
            @{ weekday = 6; available_minutes = 0; preferred_time = $null; location = $null }
        ) } | Out-Null
    Invoke-API -Method 'PUT' -Path '/v1/onboarding/cycling-context' -Body @{ cycling_context = @{
            weekly_hours = 1
            practice_duration_months = 12
            average_ride_minutes = 60
            longest_ride_minutes = 90
            weekly_rides = 2
            recent_weekly_distance_km = 40
            recent_training_weeks = 4
            training_status = 'regular'
            recent_best_distance_km = 30
            preferred_session_types = @('base')
            discipline = 'road'
            bike_type = 'road'
            terrain = 'flat'
            uses_heart_rate = $true
            uses_power = $true
            uses_gps = $true
            uses_sports_watch = $true
            uses_smart_trainer = $true
            ftp = 220
            ftp_test_date = '2026-09-10'
            ftp_protocol = '20_minute'
            average_power_watts = 185
            event_goal = $false
        } } | Out-Null

    $sql = "UPDATE users SET email_verified_at = now() WHERE email = '$email'"
    & docker compose exec -T postgres psql -X -q -v ON_ERROR_STOP=1 -U cadencia -d cadencia_dev -c $sql
    if ($LASTEXITCODE -ne 0) {
        throw 'Não foi possível preparar a conta temporária para gerar o plano.'
    }
    Write-Output 'temporary visual account verified: OK'
    if ($null -eq ($session.Cookies.GetCookies('http://localhost:8080') | Where-Object { $_.Name -eq 'cadencia_session' } | Select-Object -First 1)) {
        Invoke-API -Method 'POST' -Path '/v1/auth/login' -Body @{ email = $email; password = $password } -ExpectedStatus @(200) | Out-Null
    }
    Invoke-API -Method 'POST' -Path '/v1/plans/generate' -ExpectedStatus @(201) | Out-Null
    Write-Output 'temporary visual plan generated: OK'
    $currentPlan = Invoke-API -Method 'GET' -Path '/v1/plans/current' -ExpectedStatus @(200)
    $workoutCount = if ($null -eq $currentPlan.plan) { 0 } else { @($currentPlan.plan.workouts).Count }
    Write-Output "temporary visual current plan loaded: plan=$($null -ne $currentPlan.plan) workouts=$workoutCount"
    if ($null -ne $currentPlan.plan -and @($currentPlan.plan.workouts).Count -gt 0) {
        Write-Output ('first workout audit data: ' + (@($currentPlan.plan.workouts[0].explanation.decision_audit.data_used) -join ', '))
    }

    $cookie = $session.Cookies.GetCookies('http://localhost:8080') | Where-Object { $_.Name -eq 'cadencia_session' } | Select-Object -First 1
    if ($null -eq $cookie) {
        throw 'Cookie de sessão não encontrado para a inspeção visual.'
    }

    $chromePath = 'C:\Program Files\Google\Chrome\Application\chrome.exe'
    if (-not (Test-Path -LiteralPath $chromePath)) {
        throw 'Chrome não encontrado no caminho padrão.'
    }
    $chromeArguments = @(
        '--headless=new', '--disable-gpu', '--no-sandbox', '--no-first-run',
        "--remote-debugging-port=$cdpPort", "--user-data-dir=$chromeProfile", 'about:blank'
    )
    $chromeProcess = Start-Process -FilePath $chromePath -ArgumentList $chromeArguments -WindowStyle Hidden -PassThru

    $version = $null
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        try {
            $version = Invoke-RestMethod -UseBasicParsing "http://127.0.0.1:$cdpPort/json/version"
            break
        } catch {
            Start-Sleep -Milliseconds 250
        }
    }
    if ($null -eq $version) {
        throw 'Chrome não abriu a porta de depuração.'
    }

    $targets = Invoke-RestMethod -UseBasicParsing "http://127.0.0.1:$cdpPort/json/list"
    $target = $targets | Where-Object { $_.type -eq 'page' } | Select-Object -First 1
    if ($null -eq $target) {
        throw 'Nenhuma página CDP foi encontrada.'
    }
    $webSocket = [Net.WebSockets.ClientWebSocket]::new()
    [void]$webSocket.ConnectAsync($target.webSocketDebuggerUrl, [Threading.CancellationToken]::None).GetAwaiter().GetResult()
    Write-Output 'Chrome CDP connected: OK'
    Invoke-CDP -Method 'Network.enable' | Out-Null
    Invoke-CDP -Method 'Page.enable' | Out-Null
    Invoke-CDP -Method 'Runtime.enable' | Out-Null
    Invoke-CDP -Method 'Network.setCookie' -Params @{
        name = 'cadencia_session'
        value = $cookie.Value
        url = "$frontendURL/"
        httpOnly = $true
        sameSite = 'Lax'
    } | Out-Null

    $profileMobilePath = Join-Path $screenshots 'perfil-mobile.png'
    $profileDesktopPath = Join-Path $screenshots 'perfil-desktop.png'
    $planMobilePath = Join-Path $screenshots 'plano-mobile.png'
    $evolutionMobilePath = Join-Path $screenshots 'evolucao-mobile.png'
    $profileMobile = Open-Page -Path '/perfil' -ReadyText 'Quanto tempo cabe na sua semana?' -Width 390 -Height 844 -ScreenshotPath $profileMobilePath
    Assert-Contains $profileMobile 'FTP (watts)'
    Assert-DOMSelector '[aria-label^="Horário preferido de "]'
    if ($profileMobile.Contains('Distância da prova')) {
        throw 'Campo de evento apareceu sem event_goal.'
    }
    $null = Open-Page -Path '/perfil' -ReadyText 'Quanto tempo cabe na sua semana?' -Width 1440 -Height 900 -ScreenshotPath $profileDesktopPath

    $plan = Open-Page -Path '/plano' -ReadyText 'Uma progressão que cabe na sua rotina.' -Width 390 -Height 844 -ScreenshotPath $planMobilePath
    Click-DOMSelector '.plan-week button'
    $plan = Wait-PageText -Text 'Ver detalhes da decisão'
    Click-DOMSelector '.workout-decision-audit summary'
    Scroll-DOMSelector '.workout-decision-audit'
    $plan = Wait-PageText -Text 'Horário preferido'
    $image = Invoke-CDP -Method 'Page.captureScreenshot' -Params @{ format = 'png' }
    [IO.File]::WriteAllBytes($planMobilePath, [Convert]::FromBase64String($image.result.data))
    Assert-Contains $plan 'Objetivo secundário'
    Assert-Contains $plan 'Horário preferido'
    $evolution = Open-Page -Path '/evolucao' -ReadyText 'Os primeiros dados aparecerão após seus treinos.' -Width 390 -Height 844 -ScreenshotPath $evolutionMobilePath
    Assert-Contains $evolution 'Seu histórico, com contexto.'
    Write-Output "visual smoke test: OK"
    Write-Output "screenshots: $screenshots"
}
finally {
    if ($null -ne $webSocket) {
        try { Invoke-CDP -Method 'Browser.close' | Out-Null } catch { }
        $webSocket.Dispose()
    }
    if ($null -ne $chromeProcess -and -not $chromeProcess.HasExited) {
        Stop-Process -Id $chromeProcess.Id -Force -ErrorAction SilentlyContinue
    }
    if ($registered) {
        try {
            Invoke-API -Method 'DELETE' -Path '/v1/auth/account' -Body @{ password = $password; confirmation = 'ENCERRAR CONTA' } -ExpectedStatus @(204) | Out-Null
            Write-Output 'temporary visual account deleted: OK'
        } catch {
            Write-Error "Falha ao remover a conta temporária: $($_.Exception.Message)"
        }
    }
}
