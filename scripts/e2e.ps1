[CmdletBinding()]
param(
  [ValidateSet('run', 'up', 'down', 'logs')]
  [string]$Action = 'run',
  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]]$PlaywrightArgs
)

$rootDir = Split-Path -Parent $PSScriptRoot
& node (Join-Path $rootDir 'scripts/e2e.mjs') $Action @PlaywrightArgs
exit $LASTEXITCODE
