param(
    [Parameter(Position = 0)]
    [string] $Command = "up",

    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]] $Rest
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..\..")
$AppEnvFile = if ($env:APP_ENV_FILE) { $env:APP_ENV_FILE } else { (Join-Path $RepoRoot ".env.prod") }
if (-not [System.IO.Path]::IsPathRooted($AppEnvFile)) {
    $AppEnvFile = Join-Path $RepoRoot $AppEnvFile
}
if (-not (Test-Path $AppEnvFile)) {
    $AppEnvFile = Join-Path $RepoRoot ".env"
}
$env:APP_ENV_FILE = $AppEnvFile

$ComposeArgs = @(
    "compose",
    "--env-file", $AppEnvFile,
    "-f", (Join-Path $ScriptDir "docker-compose.yml")
)

if ($Command -eq "up") {
    $ComposeArgs += @("up", "-d")
} elseif ($Command -eq "logs") {
    $ComposeArgs += @("logs", "-f")
} else {
    $ComposeArgs += $Command
}

$ComposeArgs += $Rest

& docker @ComposeArgs
exit $LASTEXITCODE
