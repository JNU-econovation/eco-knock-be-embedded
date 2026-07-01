$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..")

& (Join-Path $RepoRoot "deploy\prod\deploy.ps1") @args
exit $LASTEXITCODE
