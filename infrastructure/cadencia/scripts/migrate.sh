#!/bin/sh
set -eu

psql -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS cadencia_schema_migrations (
  filename TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

for migration in /migrations/*.up.sql; do
  filename=$(basename "$migration")
  already_applied=$(psql -tA -v ON_ERROR_STOP=1 --set=filename="$filename" <<'SQL'
SELECT EXISTS (
  SELECT 1
  FROM cadencia_schema_migrations
  WHERE filename = :'filename'
);
SQL
)

  if [ "$already_applied" = "t" ]; then
    continue
  fi

  {
    printf 'BEGIN;\n'
    cat "$migration"
    printf "\nINSERT INTO cadencia_schema_migrations (filename) VALUES ('%s');\n" "$filename"
    printf 'COMMIT;\n'
  } | psql -v ON_ERROR_STOP=1
done

psql -v ON_ERROR_STOP=1 --set=app_user="$POSTGRES_APP_USER" <<'SQL'
GRANT USAGE ON SCHEMA public TO :"app_user";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO :"app_user";
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO :"app_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO :"app_user";
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO :"app_user";
SQL
