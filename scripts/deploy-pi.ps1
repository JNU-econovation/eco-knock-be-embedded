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

Import-EnvFile ".env.deploy"

$piHost = if ($env:PI_HOST) { $env:PI_HOST } else { "" }
$piUser = if ($env:PI_USER) { $env:PI_USER } else { "pi" }
$piSSHPort = if ($env:PI_SSH_PORT) { $env:PI_SSH_PORT } else { "22" }
$piAppDir = if ($env:PI_APP_DIR) { $env:PI_APP_DIR } else { "/home/$piUser/eco-knock-be-embedded" }
$imageName = if ($env:IMAGE_NAME) { $env:IMAGE_NAME } else { "eco-knock-be-embedded:arm64" }
$dockerPlatform = if ($env:DOCKER_PLATFORM) { $env:DOCKER_PLATFORM } else { "linux/arm64" }
$composeFile = if ($env:COMPOSE_FILE) { $env:COMPOSE_FILE } else { "docker-compose.pi.yml" }

if ($piAppDir -eq "~") {
    $piAppDir = "/home/$piUser"
} elseif ($piAppDir.StartsWith("~/")) {
    $piAppDir = "/home/$piUser/" + $piAppDir.Substring(2)
}

if ([string]::IsNullOrWhiteSpace($piHost)) {
    throw "PI_HOST is required"
}

foreach ($commandName in @("docker", "ssh", "scp")) {
    if (-not (Get-Command $commandName -ErrorAction SilentlyContinue)) {
        throw "$commandName command is required"
    }
}

if (-not (Test-Path $composeFile)) {
    throw "Compose file does not exist: $composeFile"
}

Write-Host "[1/5] Building image $imageName for $dockerPlatform"
docker buildx build --platform $dockerPlatform -t $imageName --load .

Write-Host "[2/5] Preparing remote directory $piAppDir"
ssh -p $piSSHPort "$piUser@$piHost" "mkdir -p '$piAppDir'"

Write-Host "[3/5] Transferring runtime files from $composeFile"
scp -P $piSSHPort $composeFile "$($piUser)@$($piHost):$piAppDir/docker-compose.yml"
scp -P $piSSHPort ".env" "$($piUser)@$($piHost):$piAppDir/.env"

Write-Host "[4/5] Loading docker image on Raspberry Pi"
docker save $imageName | ssh -p $piSSHPort "$piUser@$piHost" "docker load"

Write-Host "[5/5] Restarting service"
ssh -p $piSSHPort "$piUser@$piHost" "cd '$piAppDir' && docker compose up -d"

Write-Host "Deployment completed"
