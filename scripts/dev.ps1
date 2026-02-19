param(
    [ValidateSet('start', 'postgres', 'pgadmin', 'backend', 'frontend', 'seed', 'reset-db', 'reseed', 'stop', 'help')]
    [string]$Action = 'start',
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

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
  start      Start postgres+pgAdmin, then backend and frontend in new terminals
  postgres   Start only postgres + pgAdmin
  pgadmin    Start only pgAdmin
  backend    Start only backend (new terminal)
  frontend   Start only frontend (new terminal)
  seed       Run backend seeder once
  reset-db   Drop & recreate public schema (deletes all data)
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
        Start-Backend
    }
    'frontend' {
        Start-Frontend
    }
    'seed' {
        Run-Seed
    }
    'reset-db' {
        Reset-Database
    }
    'reseed' {
        Reset-Database
        Run-Seed
    }
    'stop' {
        Stop-All
    }
    'help' {
        Show-Help
    }
}
