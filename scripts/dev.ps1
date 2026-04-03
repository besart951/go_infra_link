param(
    [ValidateSet('start', 'postgres', 'pgadmin', 'backend', 'frontend', 'bootstrap', 'seed', 'reset-db', 'reseed', 'stop', 'help')]
    [string]$Action = 'start',
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

function Import-DotEnv {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        return
    }

    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if (-not $line -or $line.StartsWith('#')) {
            return
        }

        $eq = $line.IndexOf('=')
        if ($eq -lt 1) {
            return
        }

        $key = $line.Substring(0, $eq).Trim()
        $raw = $line.Substring($eq + 1).Trim()
        $hash = $raw.IndexOf(' #')
        $value = if ($hash -ge 0) { $raw.Substring(0, $hash).Trim() } else { $raw }

        if ($key) {
            [Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
}

Import-DotEnv (Join-Path $RepoRoot '.env')

function Write-Step {
    param([string]$Message)
    Write-Host "[dev] $Message" -ForegroundColor Cyan
}

function Start-Postgres {
    Write-Step 'Starting postgres + pgAdmin via docker compose...'
    Push-Location $RepoRoot
    try {
        docker compose up -d postgres pgadmin
    }
    finally {
        Pop-Location
    }
}

function Wait-ForPostgres {
    param([int]$Attempts = 30, [int]$DelaySeconds = 2)

    Write-Step 'Waiting for postgres to become ready...'
    Push-Location $RepoRoot
    try {
        for ($i = 1; $i -le $Attempts; $i++) {
            docker compose exec -T postgres sh -lc 'pg_isready -U "${POSTGRES_USER:-postgres}" -p "${POSTGRES_CONTAINER_PORT:-5432}"' *> $null
            if ($LASTEXITCODE -eq 0) {
                return
            }

            Start-Sleep -Seconds $DelaySeconds
        }
    }
    finally {
        Pop-Location
    }

    throw 'PostgreSQL did not become ready in time.'
}

function Run-DbBootstrap {
    Write-Step 'Running database bootstrap...'
    Push-Location (Join-Path $RepoRoot 'backend')
    try {
        go run .\cmd\db-bootstrap\
    }
    finally {
        Pop-Location
    }
}

function Start-Backend {
    Write-Step 'Starting backend in a new terminal...'
    Start-Process pwsh -ArgumentList '-NoExit', '-Command', "Set-Location '$RepoRoot/backend'; go run .\cmd\app\"
}

function Start-Frontend {
    Write-Step 'Starting frontend in a new terminal...'
    Start-Process pwsh -ArgumentList '-NoExit', '-Command', "Set-Location '$RepoRoot/frontend'; pnpm dev"
}

function Run-Seed {
    Write-Step 'Running seeder...'
    Push-Location (Join-Path $RepoRoot 'backend')
    try {
        go run .\cmd\seeder\
    }
    finally {
        Pop-Location
    }
}

function Reset-Database {
    if (-not $Force) {
        Write-Host ''
        Write-Host 'WARNING: This will DELETE ALL DATA in PostgreSQL.' -ForegroundColor Yellow
        $confirm = Read-Host 'Type RESET to continue'
        if ($confirm -ne 'RESET') {
            Write-Step 'Cancelled reset.'
            return
        }
    }

    Write-Step 'Resetting PostgreSQL schema public...'
    Push-Location $RepoRoot
    try {
        docker compose exec -T postgres sh -lc 'psql -U "${POSTGRES_USER:-postgres}" -d "${POSTGRES_DB:-go_infra_link}" -v ON_ERROR_STOP=1 -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO ${POSTGRES_USER:-postgres}; GRANT ALL ON SCHEMA public TO public;"'
    }
    finally {
        Pop-Location
    }
}

function Stop-All {
    Write-Step 'Stopping containers...'
    Push-Location $RepoRoot
    try {
        docker compose stop
    }
    finally {
        Pop-Location
    }
}

function Show-Help {
    @"
Usage:
  ./scripts/dev.ps1 <action> [-Force]

Actions:
  start      Start postgres+pgAdmin, run db bootstrap, then backend and frontend in new terminals
  postgres   Start only postgres + pgAdmin
  pgadmin    Start only pgAdmin
  backend    Run db bootstrap, then start only backend (new terminal)
  frontend   Start only frontend (new terminal)
  bootstrap  Run backend db bootstrap once
  seed       Run db bootstrap, then backend seeder once
  reset-db   Drop & recreate public schema, then run db bootstrap
  reseed     reset-db + seed
  stop       Stop docker compose services
  help       Show this help

Examples:
  ./scripts/dev.ps1 start
  ./scripts/dev.ps1 reset-db -Force
  ./scripts/dev.ps1 reseed -Force
"@ | Write-Host
}

switch ($Action) {
    'start' {
        Start-Postgres
        Wait-ForPostgres
        Run-DbBootstrap
        Start-Backend
        Start-Frontend
    }
    'postgres' {
        Start-Postgres
    }
    'pgadmin' {
        Write-Step 'Starting pgAdmin via docker compose...'
        Push-Location $RepoRoot
        try {
            docker compose up -d pgadmin
        }
        finally {
            Pop-Location
        }
    }
    'backend' {
        Run-DbBootstrap
        Start-Backend
    }
    'frontend' {
        Start-Frontend
    }
    'bootstrap' {
        Run-DbBootstrap
    }
    'seed' {
        Run-DbBootstrap
        Run-Seed
    }
    'reset-db' {
        Reset-Database
        Run-DbBootstrap
    }
    'reseed' {
        Reset-Database
        Run-DbBootstrap
        Run-Seed
    }
    'stop' {
        Stop-All
    }
    'help' {
        Show-Help
    }
}
