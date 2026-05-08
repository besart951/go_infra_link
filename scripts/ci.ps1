param(
    [ValidateSet('all', 'backend', 'frontend')]
    [string]$Target = 'all'
)

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$GoToolBin = Join-Path $RepoRoot '.cache/go-tools/bin'
$RacePackages = @(
    './internal/app',
    './internal/handler/middleware',
    './internal/infrastructure/exporting',
    './internal/infrastructure/realtime',
    './internal/service/exporting',
    './internal/service/notification'
)

function Write-Step {
    param([string]$Message)
    Write-Host "[ci] $Message" -ForegroundColor Green
}

function Install-GoTool {
    param(
        [string]$Binary,
        [string]$Package
    )

    New-Item -ItemType Directory -Force -Path $GoToolBin | Out-Null
    $binaryPath = Join-Path $GoToolBin $Binary
    if (-not (Test-Path $binaryPath)) {
        Write-Step "installing $Package"
        $oldGoBin = $env:GOBIN
        try {
            $env:GOBIN = $GoToolBin
            go install $Package
        }
        finally {
            $env:GOBIN = $oldGoBin
        }
    }
    return $binaryPath
}

function Invoke-BackendCI {
    Push-Location (Join-Path $RepoRoot 'backend')
    try {
        Write-Step 'backend: tests'
        go test ./...

        Write-Step 'backend: race tests'
        go test -race @RacePackages

        Write-Step 'backend: go vet'
        go vet ./...

        $staticcheck = Install-GoTool 'staticcheck.exe' 'honnef.co/go/tools/cmd/staticcheck@v0.7.0'
        Write-Step 'backend: staticcheck'
        & $staticcheck ./...

        $govulncheck = Install-GoTool 'govulncheck.exe' 'golang.org/x/vuln/cmd/govulncheck@v1.1.4'
        Write-Step 'backend: govulncheck'
        & $govulncheck ./...
    }
    finally {
        Pop-Location
    }
}

function Enable-PinnedPnpm {
    $node = Get-Command node -ErrorAction SilentlyContinue
    if (-not $node) {
        throw 'Node.js 24.x is required for frontend CI, but node was not found.'
    }

    $nodeVersion = (node -p "process.versions.node").Trim()
    $nodeMajor = [int]($nodeVersion.Split('.')[0])
    if ($nodeMajor -ne 24) {
        throw "Node.js 24.x is required for frontend CI. Current version: $nodeVersion"
    }

    $corepack = Get-Command corepack -ErrorAction SilentlyContinue
    if ($corepack) {
        corepack enable
        corepack prepare pnpm@10.29.1 --activate
    }

    $pnpmVersion = (pnpm --version).Trim()
    if ($pnpmVersion -ne '10.29.1') {
        throw "Expected pnpm 10.29.1, got $pnpmVersion"
    }
}

function Invoke-FrontendCI {
    Push-Location (Join-Path $RepoRoot 'frontend')
    try {
        Enable-PinnedPnpm

        Write-Step 'frontend: install'
        $env:CI = 'true'
        pnpm install --frozen-lockfile

        Write-Step 'frontend: check'
        pnpm check

        Write-Step 'frontend: test'
        pnpm test

        Write-Step 'frontend: build'
        pnpm build

        Write-Step 'frontend: lint'
        pnpm lint
    }
    finally {
        Pop-Location
    }
}

if ($Target -eq 'all' -or $Target -eq 'backend') {
    Invoke-BackendCI
}

if ($Target -eq 'all' -or $Target -eq 'frontend') {
    Invoke-FrontendCI
}
