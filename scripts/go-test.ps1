param([Parameter(ValueFromRemainingArguments = $true)][string[]]$GoTestArgs)

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptRoot
$runner = Join-Path $scriptRoot "go-test-exec.ps1"

if (-not (Test-Path $runner)) {
    throw "missing test runner: $runner"
}

$resolvedArgs = if ($GoTestArgs -and $GoTestArgs.Length -gt 0) { $GoTestArgs } else { @("./...") }

Push-Location $repoRoot
try {
    & go test -exec "powershell -ExecutionPolicy Bypass -File $runner" @resolvedArgs
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
