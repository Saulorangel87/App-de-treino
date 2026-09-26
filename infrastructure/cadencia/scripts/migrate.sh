#!/bin/sh
# Aplica as migrações *.up.sql pendentes, em ordem, registrando cada uma em
# cadencia_schema_migrations.
#
# Variáveis (além das PG* usadas pelo psql):
#   MIGRATIONS_DIR      diretório das migrações (padrão: /migrations)
#   DRY_RUN=1           apenas lista as migrações pendentes; não altera o banco
#   MIGRATE_STRICT=1    falha se uma migração já aplicada tiver sido modificada
#   POSTGRES_APP_USER   se definido, concede ao usuário da API as permissões
set -eu

migrations_dir="${MIGRATIONS_DIR:-/migrations}"
dry_run="${DRY_RUN:-0}"
strict="${MIGRATE_STRICT:-0}"

if [ "$dry_run" != "1" ]; then
  psql -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS cadencia_schema_migrations (
  filename TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE cadencia_schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT;
SQL
fi

table_exists=$(psql -tA -v ON_ERROR_STOP=1 -c "SELECT to_regclass('cadencia_schema_migrations') IS NOT NULL")
pending=0

for migration in "$migrations_dir"/*.up.sql; do
  filename=$(basename "$migration")
  checksum=$(sha256sum "$migration" | cut -d' ' -f1)

  if [ "$table_exists" = "t" ]; then
    recorded=$(psql -tA -v ON_ERROR_STOP=1 --set=filename="$filename" <<'SQL'
SELECT 'applied:' || COALESCE(checksum, '')
FROM cadencia_schema_migrations
WHERE filename = :'filename';
SQL
)
  else
    recorded=""
  fi

  if [ -n "$recorded" ]; then
    stored_checksum=${recorded#applied:}
    if [ -n "$stored_checksum" ] && [ "$stored_checksum" != "$checksum" ]; then
      echo "AVISO: $filename foi modificada depois de aplicada (checksum diferente)." >&2
      [ "$strict" = "1" ] && exit 1
    fi
    continue
  fi

  pending=$((pending + 1))
  if [ "$dry_run" = "1" ]; then
    echo "pendente: $filename"
    continue
  fi

  echo "aplicando: $filename"
  # O lock consultivo de transação impede que duas execuções apliquem a mesma
  # migração ao mesmo tempo; ele é liberado no COMMIT/ROLLBACK. A segunda
  # execução falha na chave primária do registro e nada é gravado.
  {
    printf 'BEGIN;\n'
    printf 'SELECT pg_advisory_xact_lock(727274);\n'
    cat "$migration"
    printf "\nINSERT INTO cadencia_schema_migrations (filename, checksum) VALUES ('%s', '%s');\n" "$filename" "$checksum"
    printf 'COMMIT;\n'
  } | psql -v ON_ERROR_STOP=1
done

if [ "$dry_run" = "1" ]; then
  echo "$pending migração(ões) pendente(s)."
  exit 0
fi

if [ -n "${POSTGRES_APP_USER:-}" ]; then
  psql -v ON_ERROR_STOP=1 --set=app_user="$POSTGRES_APP_USER" <<'SQL'
GRANT USAGE ON SCHEMA public TO :"app_user";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO :"app_user";
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO :"app_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO :"app_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO :"app_user";
SQL
fi

echo "Migrações concluídas ($pending aplicada(s))."
