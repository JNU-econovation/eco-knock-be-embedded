---
name: raspberry-pi-docker-deploy
description: Deploy this repository to a Raspberry Pi with Docker and verify runtime readiness for BME680 over I2C. Use when the user asks to deploy or redeploy the embedded server to a Raspberry Pi, prepare Docker-based Raspberry Pi execution, verify `.env.deploy` or `.env`, run `scripts/deploy-pi.ps1` or `scripts/deploy-pi.sh`, or diagnose Raspberry Pi issues around Docker, `/dev/i2c-1`, `i2cdetect`, or BME680 detection.
---

# Raspberry Pi Docker Deploy

Use this skill for this repository's Raspberry Pi deployment flow.

## Workflow

1. Read the files that drive deployment:
   - `scripts/deploy-pi.ps1`
   - `scripts/deploy-pi.sh`
   - `.env.deploy`
   - `.env`
   - `docker-compose.yml`
2. Confirm the target platform before deploying.
   - Default in this repository is `linux/arm64`.
   - On the Pi, `uname -m` must be `aarch64`. If not, stop and adjust the platform instead of forcing the deploy.
3. Prefer the platform-native deploy script.
   - On Windows, prefer `.\scripts\deploy-pi.ps1`.
   - On shell environments, prefer `./scripts/deploy-pi.sh`.
4. Treat the Pi as a runtime target, not a build target.
   - Build the image on the developer machine.
   - Load or pull the image on the Pi.
   - Do not switch back to `docker build` on the Pi unless the user explicitly asks.

## Required Inputs

Check these values before running deployment:

- `.env.deploy`
  - `PI_HOST`
  - `PI_USER`
  - `PI_SSH_PORT`
  - `PI_APP_DIR`
  - `IMAGE_NAME`
  - `DOCKER_PLATFORM`
- `.env`
  - `SERVER_HTTP_PORT`
  - `CENTRAL_BACKEND_HOST`
  - `CENTRAL_BACKEND_GRPC_PORT`
  - `ALLOW_CENTRAL_BACKEND_FAILURE`
  - `SENSOR_I2C_DEVICE`
  - `SENSOR_I2C_ADDRESS`

If `CENTRAL_BACKEND_HOST=127.0.0.1`, assume that only makes sense when the backend runs on the Pi itself. Otherwise require the real backend IP or hostname.

## Deploy

Use the repository scripts instead of rebuilding the flow by hand.

Windows:

```powershell
.\scripts\deploy-pi.ps1
```

Shell:

```bash
./scripts/deploy-pi.sh
```

These scripts are expected to:

- build `eco-knock-be-embedded:arm64`
- copy `docker-compose.yml` and `.env`
- transfer the image with `docker save | docker load`
- run `docker compose up -d`

## Verify

After deployment, verify on the Pi:

```bash
docker ps
cd ~/eco-knock-be-embedded
docker compose logs --tail=100
```

For sensor readiness, verify on the Pi host:

```bash
ls -l /dev/i2c*
sudo i2cdetect -y 1
```

Expected result for this repository:

- `/dev/i2c-1` exists
- `i2cdetect -y 1` shows `0x76` when `SENSOR_I2C_ADDRESS=0x76`
- `docker-compose.yml` exposes `/dev/i2c-1:/dev/i2c-1`

## BME680 Checks

When the container starts but the sensor is missing, check in this order:

1. Confirm the host sees `/dev/i2c-1`.
2. Confirm the host scan shows `0x76` or `0x77`.
3. Confirm `docker-compose.yml` mounts `/dev/i2c-1`.
4. Confirm the runtime config matches:
   - `SENSOR_I2C_DEVICE=/dev/i2c-1`
   - `SENSOR_I2C_ADDRESS=0x76` or `0x77`

If the host itself cannot see `0x76` or `0x77`, stop debugging the container and switch to host-side I2C troubleshooting.

## Troubleshooting

Read `references/pi-troubleshooting.md` when:

- Docker is not installed or not starting after reboot
- the root filesystem comes up read-only
- `/boot/firmware` is not mounted
- `dtparam=i2c_arm=on` looks set but `/dev/i2c-1` is still missing
- the Pi shows only `i2c-2` and GPIO 2/3 are not in `SDA1` / `SCL1` mode
