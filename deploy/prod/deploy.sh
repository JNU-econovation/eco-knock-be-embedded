#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../.." && pwd)"

if [ -f "$repo_root/.env.deploy" ]; then
  set -a
  . "$repo_root/.env.deploy"
  set +a
fi

PI_HOST="${PI_HOST:-}"
PI_USER="${PI_USER:-pi}"
PI_SSH_PORT="${PI_SSH_PORT:-22}"
PI_APP_DIR="${PI_APP_DIR:-/home/${PI_USER}/eco-knock-be-embedded}"
COMPOSE_FILE="${COMPOSE_FILE:-$script_dir/docker-compose.yml}"
APP_ENV_FILE="${APP_ENV_FILE:-.env.prod}"
IMAGE_NAME="${IMAGE_NAME:-}"
PI_PASSWORD="${PI_PASSWORD:-}"

case "${COMPOSE_FILE}" in
  /*) ;;
  *) COMPOSE_FILE="$repo_root/$COMPOSE_FILE" ;;
esac
case "${APP_ENV_FILE}" in
  /*) ;;
  *) APP_ENV_FILE="$repo_root/$APP_ENV_FILE" ;;
esac
if [ ! -f "${COMPOSE_FILE}" ] && [ "$(basename "${COMPOSE_FILE}")" = "docker-compose.pi.yml" ]; then
  COMPOSE_FILE="$script_dir/docker-compose.yml"
fi

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

if [ -n "${PI_PASSWORD}" ]; then
  if ! command -v sshpass >/dev/null 2>&1; then
    echo "sshpass command is required when PI_PASSWORD is set"
    exit 1
  fi
else
  for command_name in ssh scp; do
    if ! command -v "${command_name}" >/dev/null 2>&1; then
      echo "${command_name} command is required"
      exit 1
    fi
  done
fi

if [ ! -f "${COMPOSE_FILE}" ]; then
  echo "Compose file does not exist: ${COMPOSE_FILE}"
  exit 1
fi
if [ ! -f "${APP_ENV_FILE}" ]; then
  echo "App env file does not exist: ${APP_ENV_FILE}"
  exit 1
fi

ssh_run() {
  if [ -n "${PI_PASSWORD}" ]; then
    SSHPASS="${PI_PASSWORD}" sshpass -e ssh -p "${PI_SSH_PORT}" "${PI_USER}@${PI_HOST}" "$@"
  else
    ssh -p "${PI_SSH_PORT}" "${PI_USER}@${PI_HOST}" "$@"
  fi
}

scp_copy() {
  src="$1"
  dst="$2"

  if [ -n "${PI_PASSWORD}" ]; then
    SSHPASS="${PI_PASSWORD}" sshpass -e scp -P "${PI_SSH_PORT}" "${src}" "${PI_USER}@${PI_HOST}:${dst}"
  else
    scp -P "${PI_SSH_PORT}" "${src}" "${PI_USER}@${PI_HOST}:${dst}"
  fi
}

echo "[1/4] Preparing remote directory ${PI_APP_DIR}"
ssh_run "mkdir -p '${PI_APP_DIR}'"

echo "[2/4] Transferring runtime files from ${COMPOSE_FILE}"
scp_copy "${COMPOSE_FILE}" "${PI_APP_DIR}/docker-compose.yml"
scp_copy "${APP_ENV_FILE}" "${PI_APP_DIR}/.env"

echo "[3/4] Pulling image on Raspberry Pi"
ssh_run "cd '${PI_APP_DIR}' && IMAGE_NAME='${IMAGE_NAME}' docker compose pull"

echo "[4/4] Restarting service"
ssh_run "cd '${PI_APP_DIR}' && IMAGE_NAME='${IMAGE_NAME}' docker compose up -d"
ssh_run "cd '${PI_APP_DIR}' && IMAGE_NAME='${IMAGE_NAME}' docker compose ps"

echo "Deployment completed"
