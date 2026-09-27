#!/bin/sh
# Backup diário do PostgreSQL do Cadência.
#
# Variáveis opcionais:
#   CADENCIA_BACKUP_DIR / CADENCIA_BACKUP_RETENTION_DAYS   destino local e retenção
#   CADENCIA_OFFSITE_REMOTE   destino rclone (ex.: cadencia-crypt:backups). Se definido,
#                             o dump é copiado para fora da VPS. Use um remote "crypt":
#                             o dump contém dados pessoais e de saúde.
#   CADENCIA_OFFSITE_PRUNE=1  também apaga cópias externas antigas (exige permissão de
#                             apagar; o padrão é deixar isso a cargo do ciclo de vida do bucket).
#   CADENCIA_HEALTHCHECK_URL  URL de ping (Healthchecks.io ou similar). Recebe /start,
#                             sucesso e /fail, avisando quando o backup NÃO roda.
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
compose_file="$project_dir/infrastructure/cadencia/compose.production.yaml"
env_file="${CADENCIA_ENV_FILE:-$project_dir/infrastructure/cadencia/.env.production}"
backup_dir="${CADENCIA_BACKUP_DIR:-/var/backups/cadencia}"
retention_days="${CADENCIA_BACKUP_RETENTION_DAYS:-14}"
offsite_remote="${CADENCIA_OFFSITE_REMOTE:-}"
healthcheck_url="${CADENCIA_HEALTHCHECK_URL:-}"
timestamp=$(date -u +%Y%m%dT%H%M%SZ)

ping_monitor() {
  [ -n "$healthcheck_url" ] || return 0
  curl -fsS -m 10 --retry 3 -o /dev/null "${healthcheck_url}${1:-}" || echo "Aviso: falha ao notificar o monitor" >&2
}

finish() {
  status=$?
  if [ "$status" -ne 0 ]; then
    echo "Backup FALHOU (código $status)" >&2
    ping_monitor "/fail"
  fi
  exit "$status"
}
trap finish EXIT

ping_monitor "/start"

if [ ! -f "$env_file" ]; then
  echo "Arquivo de ambiente nao encontrado: $env_file" >&2
  exit 1
fi

umask 077
mkdir -p "$backup_dir"
backup_file="$backup_dir/cadencia-$timestamp.dump"

docker compose --env-file "$env_file" -f "$compose_file" exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --format=custom' > "$backup_file"

# Um dump vazio ou truncado também "termina" sem erro; a leitura abaixo o detecta.
[ -s "$backup_file" ] || { echo "Dump vazio: $backup_file" >&2; exit 1; }
docker run --rm -i postgres:17-alpine pg_restore --list < "$backup_file" > /dev/null
find "$backup_dir" -type f -name 'cadencia-*.dump' -mtime +"$retention_days" -delete

if [ -n "$offsite_remote" ]; then
  rclone copy "$backup_file" "$offsite_remote"
  # A retenção externa normalmente é feita por uma regra de ciclo de vida do
  # bucket, e a VM não precisa (nem deve) ter permissão de apagar backups.
  if [ "${CADENCIA_OFFSITE_PRUNE:-0}" = "1" ]; then
    rclone delete "$offsite_remote" --min-age "${retention_days}d" --include 'cadencia-*.dump'
  fi
  echo "Cópia externa concluída: $offsite_remote"
fi

echo "Backup verificado: $backup_file"
ping_monitor ""
