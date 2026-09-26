#!/usr/bin/env bash
# Aplica as migrações em um PostgreSQL descartável e executa cada fixture de
# database/tests. Todas as fixtures terminam em ROLLBACK, portanto não deixam
# dados. Usado pelo CI; também serve localmente contra um banco de teste:
#
#   PGHOST=localhost PGUSER=cadencia PGPASSWORD=... PGDATABASE=cadencia_test \
#     bash scripts/run-sql-fixtures.sh
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)

MIGRATIONS_DIR="$root/database/migrations" sh "$root/infrastructure/cadencia/scripts/migrate.sh"

failed=0
for fixture in "$root"/database/tests/*.sql; do
  name=$(basename "$fixture")
  if psql -v ON_ERROR_STOP=1 -q -f "$fixture" >/dev/null; then
    echo "ok   $name"
  else
    echo "FAIL $name" >&2
    failed=1
  fi
done

# As migrações .down.sql devem reverter na ordem inversa sem erro. Roda em
# transação descartável para não alterar o banco de verificação.
if [ "${CHECK_DOWN_MIGRATIONS:-0}" = "1" ]; then
  {
    echo "BEGIN;"
    for down in $(ls "$root"/database/migrations/*.down.sql | sort -r); do
      cat "$down"
      echo
    done
    echo "ROLLBACK;"
  } | psql -v ON_ERROR_STOP=1 -q >/dev/null && echo "ok   down migrations" || { echo "FAIL down migrations" >&2; failed=1; }
fi

exit "$failed"
