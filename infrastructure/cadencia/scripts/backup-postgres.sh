#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
compose_file="$project_dir/infrastructure/cadencia/compose.production.yaml"
env_file="${CADENCIA_ENV_FILE:-$project_dir/infrastructure/cadencia/.env.production}"
backup_dir="${CADENCIA_BACKUP_DIR:-/var/backups/cadencia}"
retention_days="${CADENCIA_BACKUP_RETENTION_DAYS:-14}"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)

if [ ! -f "$env_file" ]; then
  echo "Arquivo de ambiente nao encontrado: $env_file" >&2
  exit 1
fi

umask 077
mkdir -p "$backup_dir"
backup_file="$backup_dir/cadencia-$timestamp.dump"

docker compose --env-file "$env_file" -f "$compose_file" exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom' > "$backup_file"

docker run --rm -i postgres:17-alpine pg_restore --list < "$backup_file" > /dev/null
find "$backup_dir" -type f -name 'cadencia-*.dump' -mtime +"$retention_days" -delete

echo "Backup verificado: $backup_file"
