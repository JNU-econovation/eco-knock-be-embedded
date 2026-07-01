$ErrorActionPreference = "Stop"

function Import-EnvFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if (-not (Test-Path $Path)) {
        return
    }

    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) {
            return
        }

        $parts = $line -split "=", 2
        if ($parts.Count -ne 2) {
            return
        }

        $name = $parts[0].Trim()
        $value = $parts[1].Trim()
        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

function Escape-RemoteSingleQuoted {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Value
    )

    return $Value.Replace("'", "'\''")
}

function Test-ParamikoAvailable {
    $script = @'
import paramiko
'@

    $script | python -W ignore - 2>$null
    return $LASTEXITCODE -eq 0
}

function Invoke-PiCommand {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Command
    )

    if (-not $usePasswordAuth) {
        ssh -p $piSSHPort "$piUser@$piHost" $Command
        return
    }

    $env:DEPLOY_PI_HOST = $piHost
    $env:DEPLOY_PI_USER = $piUser
    $env:DEPLOY_PI_PASSWORD = $piPassword
    $env:DEPLOY_PI_SSH_PORT = $piSSHPort
    $env:DEPLOY_REMOTE_COMMAND = $Command

    @'
import os
import sys
import paramiko

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(
    hostname=os.environ["DEPLOY_PI_HOST"],
    port=int(os.environ["DEPLOY_PI_SSH_PORT"]),
    username=os.environ["DEPLOY_PI_USER"],
    password=os.environ["DEPLOY_PI_PASSWORD"],
    timeout=15,
    look_for_keys=False,
    allow_agent=False,
)
stdin, stdout, stderr = client.exec_command(os.environ["DEPLOY_REMOTE_COMMAND"], get_pty=True)
out = stdout.read()
err = stderr.read()
status = stdout.channel.recv_exit_status()
client.close()
sys.stdout.buffer.write(out)
sys.stderr.buffer.write(err)
sys.exit(status)
'@ | python -W ignore -

    if ($LASTEXITCODE -ne 0) {
        throw "Remote command failed: $Command"
    }
}

function Copy-ToPi {
    param(
        [Parameter(Mandatory = $true)]
        [string]$LocalPath,
        [Parameter(Mandatory = $true)]
        [string]$RemotePath
    )

    if (-not $usePasswordAuth) {
        scp -P $piSSHPort $LocalPath "$($piUser)@$($piHost):$RemotePath"
        return
    }

    $env:DEPLOY_PI_HOST = $piHost
    $env:DEPLOY_PI_USER = $piUser
    $env:DEPLOY_PI_PASSWORD = $piPassword
    $env:DEPLOY_PI_SSH_PORT = $piSSHPort
    $env:DEPLOY_LOCAL_PATH = (Resolve-Path $LocalPath).Path
    $env:DEPLOY_REMOTE_PATH = $RemotePath

    @'
import os
import paramiko

transport = paramiko.Transport((os.environ["DEPLOY_PI_HOST"], int(os.environ["DEPLOY_PI_SSH_PORT"])))
transport.connect(username=os.environ["DEPLOY_PI_USER"], password=os.environ["DEPLOY_PI_PASSWORD"])
sftp = paramiko.SFTPClient.from_transport(transport)
sftp.put(os.environ["DEPLOY_LOCAL_PATH"], os.environ["DEPLOY_REMOTE_PATH"])
sftp.close()
transport.close()
'@ | python -W ignore -

    if ($LASTEXITCODE -ne 0) {
        throw "File transfer failed: $LocalPath -> $RemotePath"
    }
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..\..")

Import-EnvFile (Join-Path $RepoRoot ".env.deploy")

$piHost = if ($env:PI_HOST) { $env:PI_HOST } else { "" }
$piUser = if ($env:PI_USER) { $env:PI_USER } else { "pi" }
$piSSHPort = if ($env:PI_SSH_PORT) { $env:PI_SSH_PORT } else { "22" }
$piAppDir = if ($env:PI_APP_DIR) { $env:PI_APP_DIR } else { "/home/$piUser/eco-knock-be-embedded" }
$composeFile = if ($env:COMPOSE_FILE) { $env:COMPOSE_FILE } else { (Join-Path $ScriptDir "docker-compose.yml") }
$appEnvFile = if ($env:APP_ENV_FILE) { $env:APP_ENV_FILE } else { ".env.prod" }
$imageName = if ($env:IMAGE_NAME) { $env:IMAGE_NAME } else { "" }
$piPassword = if ($env:PI_PASSWORD) { $env:PI_PASSWORD } else { "" }
$usePasswordAuth = -not [string]::IsNullOrWhiteSpace($piPassword)

if (-not [System.IO.Path]::IsPathRooted($composeFile)) {
    $composeFile = Join-Path $RepoRoot $composeFile
}
if (-not [System.IO.Path]::IsPathRooted($appEnvFile)) {
    $appEnvFile = Join-Path $RepoRoot $appEnvFile
}
if (-not (Test-Path $composeFile) -and (Split-Path -Leaf $composeFile) -eq "docker-compose.pi.yml") {
    $composeFile = Join-Path $ScriptDir "docker-compose.yml"
}

if ($piAppDir -eq "~") {
    $piAppDir = "/home/$piUser"
} elseif ($piAppDir.StartsWith("~/")) {
    $piAppDir = "/home/$piUser/" + $piAppDir.Substring(2)
}

if ([string]::IsNullOrWhiteSpace($piHost)) {
    throw "PI_HOST is required"
}

if ($usePasswordAuth) {
    if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
        throw "python command is required when PI_PASSWORD is set"
    }
    if (-not (Test-ParamikoAvailable)) {
        throw "Python package paramiko is required when PI_PASSWORD is set"
    }
} else {
    foreach ($commandName in @("ssh", "scp")) {
        if (-not (Get-Command $commandName -ErrorAction SilentlyContinue)) {
            throw "$commandName command is required"
        }
    }
}

if (-not (Test-Path $composeFile)) {
    throw "Compose file does not exist: $composeFile"
}
if (-not (Test-Path $appEnvFile)) {
    throw "App env file does not exist: $appEnvFile"
}

$escapedPiAppDir = Escape-RemoteSingleQuoted $piAppDir
$remoteComposeEnv = ""
if (-not [string]::IsNullOrWhiteSpace($imageName)) {
    $escapedImageName = Escape-RemoteSingleQuoted $imageName
    $remoteComposeEnv = "IMAGE_NAME='$escapedImageName' "
}

Write-Host "[1/4] Preparing remote directory $piAppDir"
Invoke-PiCommand "mkdir -p '$escapedPiAppDir'"

Write-Host "[2/4] Transferring runtime files from $composeFile"
Copy-ToPi $composeFile "$piAppDir/docker-compose.yml"
Copy-ToPi $appEnvFile "$piAppDir/.env"

Write-Host "[3/4] Pulling image on Raspberry Pi"
Invoke-PiCommand "cd '$escapedPiAppDir' && ${remoteComposeEnv}docker compose pull"

Write-Host "[4/4] Restarting service"
Invoke-PiCommand "cd '$escapedPiAppDir' && ${remoteComposeEnv}docker compose up -d"
Invoke-PiCommand "cd '$escapedPiAppDir' && ${remoteComposeEnv}docker compose ps"

Write-Host "Deployment completed"
