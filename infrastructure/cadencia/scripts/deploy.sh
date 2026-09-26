#!/bin/sh
# Deploy do Cadência na VPS: atualiza o checkout, faz backup, aplica migrações
# na ordem, recria os serviços e roda o smoke test. Cada etapa para no primeiro
# erro. Execute na VPS, a partir de qualquer diretório:
#
#   sh infrastructure/cadencia/scripts/deploy.sh [--yes]
#
# Smoke test autenticado (recomendado): usa uma conta de teste dedicada, com
# e-mail confirmado. As credenciais (CADENCIA_SMOKE_EMAIL e
# CADENCIA_SMOKE_PASSWORD) são lidas de /etc/cadencia/smoke.env, se existir, ou
# do ambiente. Sem elas, essa etapa é pulada e o script avisa; healthchecks
# sozinhos NÃO comprovam compatibilidade do schema.
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
cd "$project_dir"
env_file="${CADENCIA_ENV_FILE:-$project_dir/infrastructure/cadencia/.env.production}"
compose="docker compose --env-file $env_file -f $project_dir/infrastructure/cadencia/compose.production.yaml"
frontend_url="${CADENCIA_FRONTEND_URL:-https://cadencia.devsaulo.com.br}"
api_url="${CADENCIA_API_URL:-https://cadencia-api.devsaulo.com.br}"
assume_yes=0
[ "${1:-}" = "--yes" ] && assume_yes=1

step() { printf '\n==> %s\n' "$1"; }
fail() { printf 'ERRO: %s\n' "$1" >&2; exit 1; }

[ -f "$env_file" ] || fail "arquivo de ambiente ausente: $env_file"
[ -z "$(git status --porcelain)" ] || fail "o checkout da VPS tem alterações locais; resolva antes do deploy"

step "Atualizando o checkout (fast-forward)"
before=$(git rev-parse --short HEAD)
git fetch origin master
git log --oneline "HEAD..origin/master"
if [ "$assume_yes" -ne 1 ]; then
  printf 'Continuar com o deploy? [s/N] '
  read -r answer
  [ "$answer" = "s" ] || [ "$answer" = "S" ] || fail "deploy cancelado"
fi
git merge --ff-only origin/master
after=$(git rev-parse --short HEAD)
echo "commit: $before -> $after"

step "Backup preventivo"
CADENCIA_ENV_FILE="$env_file" sh infrastructure/cadencia/scripts/backup-postgres.sh

step "Build das imagens"
$compose build

step "PostgreSQL e migrações pendentes"
$compose up -d postgres
$compose --profile maintenance run --rm -e DRY_RUN=1 migrate
$compose --profile maintenance run --rm migrate

step "Recriando API, frontend e tunnel"
$compose up -d api frontend tunnel

step "Aguardando /ready da API (inclui verificação do schema)"
ready=0
for _ in $(seq 1 30); do
  if $compose exec -T api wget -qO- http://127.0.0.1:8080/ready | grep -q '"ready"'; then ready=1; break; fi
  sleep 3
done
[ "$ready" -eq 1 ] || { $compose logs --tail 50 api; fail "a API não ficou pronta; considere restaurar o backup"; }

# O tunnel acabou de ser recriado e leva alguns segundos para reconectar; até lá
# a Cloudflare responde 530. Tenta por até ~90 s antes de falhar.
wait_public() {
  label="$1"; url="$2"; pattern="${3:-}"
  for _ in $(seq 1 30); do
    if body=$(curl -fsS -m 15 "$url" 2>/dev/null) && { [ -z "$pattern" ] || printf '%s' "$body" | grep -q "$pattern"; }; then
      echo "$label ok"; return 0
    fi
    sleep 3
  done
  fail "$label não respondeu em $url"
}

step "Smoke test público"
wait_public "api /health" "$api_url/health"
wait_public "api /ready" "$api_url/ready" '"ready"'
wait_public "frontend" "$frontend_url/"

step "Smoke test autenticado (/v1/plans/current)"
smoke_env="${CADENCIA_SMOKE_ENV_FILE:-/etc/cadencia/smoke.env}"
if [ -z "${CADENCIA_SMOKE_EMAIL:-}" ] && [ -r "$smoke_env" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$smoke_env"
  set +a
fi
if [ -n "${CADENCIA_SMOKE_EMAIL:-}" ] && [ -n "${CADENCIA_SMOKE_PASSWORD:-}" ]; then
  jar=$(mktemp)
  trap 'rm -f "$jar"' EXIT
  curl -fsS -m 15 -c "$jar" -H "Origin: $frontend_url" -H 'Content-Type: application/json' \
    -d "{\"email\":\"$CADENCIA_SMOKE_EMAIL\",\"password\":\"$CADENCIA_SMOKE_PASSWORD\"}" \
    "$api_url/v1/auth/login" -o /dev/null
  code=$(curl -sS -m 15 -b "$jar" -o /dev/null -w '%{http_code}' "$api_url/v1/plans/current")
  curl -sS -m 15 -b "$jar" -H "Origin: $frontend_url" -X POST "$api_url/v1/auth/logout" -o /dev/null || true
  # 200 (com ou sem plano) prova que o schema atende ao código; 500 indica defasagem.
  case "$code" in
    200) echo "GET /v1/plans/current -> $code ok" ;;
    *) fail "GET /v1/plans/current retornou $code" ;;
  esac
else
  echo "PULADO: defina CADENCIA_SMOKE_EMAIL e CADENCIA_SMOKE_PASSWORD para validar uma rota autenticada."
fi

step "Concluído"
$compose ps
echo "Registre o deploy em docs/STATUS.md (commit $after)."
