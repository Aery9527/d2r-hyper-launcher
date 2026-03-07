param([Parameter(ValueFromRemainingArguments = $true)][string[]]$argv)

if (-not $argv -or $argv.Length -eq 0) {
    throw "missing test executable path"
}

$exe = $argv[0]
$rest = if ($argv.Length -gt 1) { $argv[1..($argv.Length - 1)] } else { @() }

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptRoot
$targetDir = Join-Path $repoRoot ".tmp\go-test-runner"
$tmpRoot = Split-Path -Parent $targetDir

New-Item -ItemType Directory -Force -Path $targetDir | Out-Null

$base = [System.IO.Path]::GetFileName($exe)
$target = Join-Path $targetDir $base

Copy-Item -Force $exe $target
& $target @rest
$code = $LASTEXITCODE
Remove-Item -Force $target -ErrorAction SilentlyContinue
if (Test-Path $targetDir) {
    $remaining = Get-ChildItem $targetDir -Force -ErrorAction SilentlyContinue
    if (-not $remaining) {
        Remove-Item -Force $targetDir -ErrorAction SilentlyContinue
    }
}
if (Test-Path $tmpRoot) {
    $remaining = Get-ChildItem $tmpRoot -Force -ErrorAction SilentlyContinue
    if (-not $remaining) {
        Remove-Item -Force $tmpRoot -ErrorAction SilentlyContinue
    }
}

exit $code
