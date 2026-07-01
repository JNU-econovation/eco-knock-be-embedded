---
name: raspberry-pi-docker-deploy
description: Deploy this repository to a Raspberry Pi with Docker and verify runtime readiness for sensors, GHCR images, and exposed ports. Use when the user asks to deploy or redeploy the embedded server to a Raspberry Pi, prepare Docker-based Raspberry Pi execution, verify `.env.deploy` or the selected app env file, run `deploy/prod/deploy.ps1`, `deploy/prod/deploy.sh`, `scripts/deploy-pi.ps1`, or `scripts/deploy-pi.sh`, or diagnose Raspberry Pi issues around GHCR auth, Docker compose, allocated ports, `/dev/i2c-1`, `i2cdetect`, BME680, or VEML7700 detection.
---

# Raspberry Pi Docker Deploy

Use this skill for this repository's Raspberry Pi deployment flow. The current default flow is GHCR pull on the Pi, not local `docker save` / `docker load`.

## Workflow

1. Read the files that drive deployment:
   - `deploy/prod/deploy.ps1`
   - `deploy/prod/deploy.sh`
   - `deploy/prod/prod.ps1`
   - `deploy/prod/prod.sh`
   - `deploy/prod/docker-compose.yml`
   - `scripts/deploy-pi.ps1`
   - `scripts/deploy-pi.sh`
   - `.env.deploy`
   - app env file selected by `APP_ENV_FILE`, default `.env.prod`
2. Confirm the target platform before deploying.
   - Default in this repository is `linux/arm64`.
   - On the Pi, `uname -m` must be `aarch64`. If not, stop and adjust the platform instead of forcing the deploy.
3. Prefer the platform-native deploy script.
   - On Windows, prefer `.\deploy\prod\deploy.ps1`.
   - On shell environments, prefer `./deploy/prod/deploy.sh`.
   - `.\scripts\deploy-pi.ps1` and `./scripts/deploy-pi.sh` are compatibility wrappers.
4. Treat the Pi as a runtime target, not a build target.
   - GitHub Actions builds and pushes the `linux/arm64` image to GHCR.
   - The Pi pulls the image with `docker compose pull`.
   - Do not switch back to `docker build` on the Pi unless the user explicitly asks.

## Required Inputs

Check these values before running deployment:

- `.env.deploy`
  - `PI_HOST`
  - `PI_USER`
  - `PI_SSH_PORT`
  - `PI_APP_DIR`
  - `IMAGE_NAME`
  - `COMPOSE_FILE`
  - `APP_ENV_FILE`
  - `PI_PASSWORD` when password-based SSH is expected
- selected app env file, usually `.env.prod`
  - `SERVER_HTTP_PORT`
  - `SERVER_GRPC_PORT`
  - `CENTRAL_BACKEND_HOST`
  - `CENTRAL_BACKEND_GRPC_PORT`
  - `ALLOW_CENTRAL_BACKEND_FAILURE`
  - `SENSOR_READER_MODE`
  - `LIGHT_SENSOR_READER_MODE`
  - `AIR_PURIFIER_CLIENT_MODE`
  - `SENSOR_I2C_DEVICE`
  - `SENSOR_I2C_ADDRESS`
  - `LIGHT_SENSOR_I2C_DEVICE`
  - `LIGHT_SENSOR_I2C_ADDRESS`

If `CENTRAL_BACKEND_HOST=127.0.0.1`, assume that only makes sense when the backend runs on the Pi itself. Otherwise require the real backend IP or hostname.
If `APP_ENV_FILE` is absent, deployment defaults to `.env.prod`. The deploy script copies that file to the Pi as `.env`.
`DOCKER_PLATFORM` is legacy/local-build context only; it is not used by the GHCR pull deployment path.

## Deploy

Use the repository scripts instead of rebuilding the flow by hand.

Windows:

```powershell
.\deploy\prod\deploy.ps1
```

Shell:

```bash
./deploy/prod/deploy.sh
```

These scripts are expected to:

- read `.env.deploy`
- copy `deploy/prod/docker-compose.yml` to the Pi as `docker-compose.yml`
- copy `APP_ENV_FILE` to the Pi as `.env`
- run `docker compose pull`
- run `docker compose up -d`
- print `docker compose ps`

If GHCR returns `unauthorized`, the package is private or the Pi is not logged in. Fix with one of:

```bash
docker login ghcr.io
```

or make the GHCR package public when organization policy allows it.

## Verify

After deployment, verify on the Pi:

```bash
docker ps
cd ~/eco-knock-be-embedded
docker compose logs --tail=100
```

From the developer PC, verify exposed ports:

```powershell
Test-NetConnection <PI_HOST> -Port 6565
Test-NetConnection <PI_HOST> -Port 19090
```

For sensor readiness, verify on the Pi host:

```bash
ls -l /dev/i2c*
sudo i2cdetect -y 1
```

Expected result for this repository:

- `/dev/i2c-1` exists
- `i2cdetect -y 1` shows `76` or `77` for BME680 when the real sensor is attached
- `i2cdetect -y 1` shows `10` for VEML7700 when the real light sensor is attached
- `docker-compose.yml` exposes `/dev/i2c-1:/dev/i2c-1`
- `SERVER_HTTP_PORT` and `SERVER_GRPC_PORT` are published by the server container

## BME680 Checks

When the container starts but the sensor is missing, check in this order:

1. Confirm the host sees `/dev/i2c-1`.
2. Confirm the host scan shows `0x76` or `0x77`.
3. Confirm `docker-compose.yml` mounts `/dev/i2c-1`.
4. Confirm the runtime config matches:
   - `SENSOR_READER_MODE=real`
   - `SENSOR_I2C_DEVICE=/dev/i2c-1`
   - `SENSOR_I2C_ADDRESS=0x76` or `0x77`

If the host itself cannot see `0x76` or `0x77`, stop debugging the container and switch to host-side I2C troubleshooting.

## VEML7700 Checks

When the light sensor fails in real mode, check in this order:

1. Confirm `LIGHT_SENSOR_READER_MODE=real`.
2. Confirm the host scan shows `0x10`.
3. Confirm runtime config matches:
   - `LIGHT_SENSOR_I2C_DEVICE=/dev/i2c-1`
   - `LIGHT_SENSOR_I2C_ADDRESS=0x10`
4. If the host scan does not show `0x10`, debug wiring and I2C on the Pi host before debugging the container.

## Port Conflict Checks

If deployment fails with `port is already allocated`, check for old compose projects or containers:

```bash
docker compose ls
docker ps --format 'table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.Names}}'
ss -ltnp | grep -E ':6565|:19090' || true
```

When migrating from the old compose project, remove both the old and failed new project before bringing the new one up:

```bash
cd ~/eco-knock-be-embedded
docker compose -p eco-knock-be-embedded down --remove-orphans
docker compose -p eco-knock-embedded-prod down --remove-orphans
IMAGE_NAME=ghcr.io/jnu-econovation/eco-knock-be-embedded:latest docker compose up -d
```

## Troubleshooting

Read `references/pi-troubleshooting.md` when:

- Docker is not installed or not starting after reboot
- the root filesystem comes up read-only
- `/boot/firmware` is not mounted
- `dtparam=i2c_arm=on` looks set but `/dev/i2c-1` is still missing
- the Pi shows only `i2c-2` and GPIO 2/3 are not in `SDA1` / `SCL1` mode
