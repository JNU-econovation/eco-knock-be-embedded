#!/usr/bin/env sh

set -eu

if [ -f ".env.deploy" ]; then
  set -a
  . ./.env.deploy
  set +a
fi

PI_HOST="${PI_HOST:-}"
PI_USER="${PI_USER:-pi}"
PI_SSH_PORT="${PI_SSH_PORT:-22}"
PI_APP_DIR="${PI_APP_DIR:-/home/${PI_USER}/eco-knock-be-embedded}"
IMAGE_NAME="${IMAGE_NAME:-eco-knock-be-embedded:arm64}"
DOCKER_PLATFORM="${DOCKER_PLATFORM:-linux/arm64}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.pi.yml}"

case "${PI_APP_DIR}" in
  "~")
    PI_APP_DIR="/home/${PI_USER}"
    ;;
  "~/"*)
    PI_APP_DIR="/home/${PI_USER}/${PI_APP_DIR#~/}"
    ;;
esac

if [ -z "${PI_HOST}" ]; then
  echo "PI_HOST is required"
  exit 1
fi

for command_name in docker ssh scp; do
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "${command_name} command is required"
    exit 1
  fi
done

if [ ! -f "${COMPOSE_FILE}" ]; then
  echo "Compose file does not exist: ${COMPOSE_FILE}"
  exit 1
fi

echo "[1/5] Building image ${IMAGE_NAME} for ${DOCKER_PLATFORM}"
docker buildx build --platform "${DOCKER_PLATFORM}" -t "${IMAGE_NAME}" --load .

echo "[2/5] Preparing remote directory ${PI_APP_DIR}"
ssh -p "${PI_SSH_PORT}" "${PI_USER}@${PI_HOST}" "mkdir -p '${PI_APP_DIR}'"

echo "[3/5] Transferring runtime files from ${COMPOSE_FILE}"
scp -P "${PI_SSH_PORT}" "${COMPOSE_FILE}" "${PI_USER}@${PI_HOST}:${PI_APP_DIR}/docker-compose.yml"
scp -P "${PI_SSH_PORT}" ".env" "${PI_USER}@${PI_HOST}:${PI_APP_DIR}/.env"

echo "[4/5] Loading docker image on Raspberry Pi"
docker save "${IMAGE_NAME}" | ssh -p "${PI_SSH_PORT}" "${PI_USER}@${PI_HOST}" "docker load"

echo "[5/5] Restarting service"
ssh -p "${PI_SSH_PORT}" "${PI_USER}@${PI_HOST}" "cd '${PI_APP_DIR}' && docker compose up -d"

echo "Deployment completed"
