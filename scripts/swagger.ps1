param()

$ErrorActionPreference = 'Stop'
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path

Push-Location (Join-Path $RepoRoot 'backend')
try {
    go run ./cmd/swagger
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
    go run ./cmd/permission-contract
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}

Push-Location (Join-Path $RepoRoot 'frontend')
try {
    pnpm run api:generate
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}
