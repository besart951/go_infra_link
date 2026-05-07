# ─────────────────────────────────────────────────────────────
#  go_infra_link — Production Deploy Script
#
#  Usage:
#    ./scripts/prod.ps1            # build images + start stack
#    ./scripts/prod.ps1 -Down      # stop and remove containers
#    ./scripts/prod.ps1 -Pull      # pull latest base images before building
#    ./scripts/prod.ps1 -Logs      # tail logs after start
#
#  Prerequisites:
#    1. .env exists with APP_ENV=production and all required values
#       (copy .env.prod.example to .env and fill in CHANGE_ME values)
#    2. Docker Desktop / Docker Engine + Compose V2 are installed
# ─────────────────────────────────────────────────────────────
param(
    [switch]$Down,
    [switch]$Pull,
    [switch]$Logs
)

$ErrorActionPreference = 'Stop'
$RepoRoot      = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$ComposeFile   = Join-Path $RepoRoot 'docker-compose.prod.yml'
$EnvFile       = Join-Path $RepoRoot '.env'

function Write-Step { param([string]$M) Write-Host "[prod] $M" -ForegroundColor Green }
function Write-Warn { param([string]$M) Write-Host "[prod] WARNING: $M" -ForegroundColor Yellow }
function Write-Fail { param([string]$M) Write-Host "[prod] ERROR: $M" -ForegroundColor Red; exit 1 }

# ── Validate prerequisites ────────────────────────────────────
if (-not (Test-Path $EnvFile)) {
    Write-Fail ".env not found. Copy .env.prod.example to .env and fill in all CHANGE_ME values."
}

function Get-EnvValue {
    param([string]$Key)
    $match = Select-String -Path $EnvFile -Pattern "^\s*$Key\s*=" | Select-Object -First 1
    if (-not $match) { return $null }
    # Strip inline comments and surrounding whitespace
    return (($match.Line -split '=', 2)[1].Trim() -replace '\s*#.*$', '').Trim()
}

$appEnv = Get-EnvValue 'APP_ENV'
if ($appEnv -ne 'production') {
    Write-Warn ".env has APP_ENV=$appEnv — expected 'production'."
    $confirm = Read-Host "Continue deploying with APP_ENV=$appEnv? (y/N)"
    if ($confirm -notmatch '^[Yy]$') {
        Write-Step "Aborted. Set APP_ENV=production in .env to deploy."
        exit 0
    }
}

# Check for unfilled CHANGE_ME placeholders in required vars
$requiredVars = @('JWT_SECRET', 'POSTGRES_USER', 'POSTGRES_PASSWORD', 'POSTGRES_DB',
                  'POSTGRES_SSLMODE', 'TRUSTED_PROXIES', 'COOKIE_SECURE',
                  'DATABASE_URL', 'SEED_USER_EMAIL', 'SEED_USER_PASSWORD')
foreach ($var in $requiredVars) {
    $val = Get-EnvValue $var
    if (-not $val) {
        Write-Fail "$var is not set in .env. See .env.prod.example."
    }
    if ($val -match 'CHANGE_ME') {
        Write-Fail "$var still contains 'CHANGE_ME'. Set a real value in .env before deploying."
    }
}

if ((Get-EnvValue 'COOKIE_SECURE') -ne 'true') {
    Write-Fail "COOKIE_SECURE must be true in production."
}

$corsAllowedOrigins = Get-EnvValue 'CORS_ALLOWED_ORIGINS'
$corsOriginValues = @()
if ($corsAllowedOrigins) {
    $corsOriginValues = ($corsAllowedOrigins -split ',') | ForEach-Object { $_.Trim() }
}
if ($corsOriginValues -contains '*') {
    Write-Fail "CORS_ALLOWED_ORIGINS must not contain wildcard '*'."
}

$sslMode = (Get-EnvValue 'POSTGRES_SSLMODE').ToLowerInvariant()
$allowUnsafeSSLMode = (Get-EnvValue 'DB_ALLOW_UNSAFE_SSLMODE')
if (($sslMode -in @('disable', 'allow', 'prefer')) -and ($allowUnsafeSSLMode -ne 'true')) {
    Write-Fail "POSTGRES_SSLMODE=$sslMode is unsafe. Use require/verify-ca/verify-full or set DB_ALLOW_UNSAFE_SSLMODE=true intentionally."
}

# ── Actions ───────────────────────────────────────────────────
Push-Location $RepoRoot
try {
    if ($Down) {
        Write-Step "Stopping and removing production containers..."
        docker compose -f $ComposeFile down
        Write-Step "Done."
        return
    }

    if ($Pull) {
        Write-Step "Pulling latest base images..."
        docker compose -f $ComposeFile pull --ignore-pull-failures 2>&1 | Out-Null
    }

    Write-Step "Building images and starting production stack..."
    docker compose -f $ComposeFile up -d --build

    Write-Step ""
    Write-Step "Production stack is running."
    Write-Step "  Backend health : http://localhost:$(Get-EnvValue 'BACKEND_PORT' ?? '8080')/health"
    Write-Step "  Frontend       : http://localhost:$(Get-EnvValue 'FRONTEND_PORT' ?? '80')"
    Write-Step ""
    Write-Step "Useful commands:"
    Write-Step "  Logs  : docker compose -f docker-compose.prod.yml logs -f"
    Write-Step "  Stop  : ./scripts/prod.ps1 -Down"
    Write-Step "  Status: docker compose -f docker-compose.prod.yml ps"

    if ($Logs) {
        docker compose -f $ComposeFile logs -f
    }
}
finally {
    Pop-Location
}
