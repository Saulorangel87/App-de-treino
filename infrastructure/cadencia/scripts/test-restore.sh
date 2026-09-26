#!/bin/sh
# Prova que um backup é restaurável: sobe um PostgreSQL descartável, restaura o
# dump e confere o registro de migrações. Não toca na produção.
#
# Uso: test-restore.sh [arquivo.dump]   (padrão: o dump mais recente)
# Opcional: CADENCIA_HEALTHCHECK_URL recebe sucesso ou /fail.
set -eu

backup_dir="${CADENCIA_BACKUP_DIR:-/var/backups/cadencia}"
healthcheck_url="${CADENCIA_HEALTHCHECK_URL:-}"
dump="${1:-$(ls -1t "$backup_dir"/cadencia-*.dump 2>/dev/null | head -n 1)}"
container="cadencia-restore-test-$$"

ping_monitor() {
  [ -n "$healthcheck_url" ] || return 0
  curl -fsS -m 10 --retry 3 -o /dev/null "${healthcheck_url}${1:-}" || true
}

cleanup() {
  status=$?
  docker rm -f "$container" >/dev/null 2>&1 || true
  if [ "$status" -ne 0 ]; then
    echo "Teste de restauração FALHOU" >&2
    ping_monitor "/fail"
  fi
  exit "$status"
}
trap cleanup EXIT

[ -n "$dump" ] && [ -s "$dump" ] || { echo "Nenhum dump encontrado em $backup_dir" >&2; exit 1; }
echo "Restaurando $dump"

docker run -d --name "$container" -e POSTGRES_PASSWORD=restore-test -e POSTGRES_DB=restore \
  postgres:17-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24 >/dev/null

for _ in $(seq 1 30); do
  docker exec "$container" pg_isready -U postgres -d restore >/dev/null 2>&1 && break
  sleep 2
done

docker exec -i "$container" pg_restore -U postgres -d restore --no-owner --no-privileges < "$dump"

query() { docker exec "$container" psql -U postgres -d restore -tA -c "$1"; }

migrations=$(query "SELECT count(*) FROM cadencia_schema_migrations")
latest=$(query "SELECT filename FROM cadencia_schema_migrations ORDER BY filename DESC LIMIT 1")
users=$(query "SELECT count(*) FROM users")
[ "$migrations" -gt 0 ] || { echo "Registro de migrações vazio" >&2; exit 1; }

echo "Restauração verificada: $migrations migrações (última: $latest), $users usuário(s)."
ping_monitor ""
