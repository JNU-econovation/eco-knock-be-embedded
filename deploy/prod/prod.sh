#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../.." && pwd)"
app_env_file="${APP_ENV_FILE:-$repo_root/.env.prod}"
case "$app_env_file" in
    /*) ;;
    *) app_env_file="$repo_root/$app_env_file" ;;
esac
if [ ! -f "$app_env_file" ]; then
    app_env_file="$repo_root/.env"
fi
export APP_ENV_FILE="$app_env_file"

command="${1:-up}"
if [ "$#" -gt 0 ]; then
    shift
fi

case "$command" in
    up)
        set -- up -d "$@"
        ;;
    logs)
        set -- logs -f "$@"
        ;;
    *)
        set -- "$command" "$@"
        ;;
esac

docker compose \
    --env-file "$app_env_file" \
    -f "$script_dir/docker-compose.yml" \
    "$@"
