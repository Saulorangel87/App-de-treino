#!/bin/sh
# Cópia completa dos dados de um atleta (LGPD, art. 18), com saúde e limitações.
# Gera um JSON na pasta do usuário, só legível por ele. Só lê o banco. Execute na
# VPS, de qualquer diretório:
#
#   sh infrastructure/cadencia/scripts/account-export.sh atleta@exemplo.com
#
# O arquivo tem dados de saúde: confirme quem pediu antes de enviar, envie
# compactado com senha e apague o arquivo da VPS depois.
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)
env_file="${CADENCIA_ENV_FILE:-$project_dir/infrastructure/cadencia/.env.production}"
compose_file="$project_dir/infrastructure/cadencia/compose.production.yaml"

fail() { printf 'ERRO: %s\n' "$1" >&2; exit 1; }

[ "$#" -eq 1 ] || fail "uso: sh infrastructure/cadencia/scripts/account-export.sh <e-mail do atleta>"
email=$(printf '%s' "$1" | tr 'A-Z' 'a-z')
case "$email" in
  *[!a-z0-9@._+-]*) fail "e-mail com caracteres inválidos" ;;
  ?*@?*.?*) ;;
  *) fail "e-mail inválido: $email" ;;
esac
[ -f "$env_file" ] || fail "arquivo de ambiente ausente: $env_file"

umask 077
safe_name=$(printf '%s' "$email" | tr '@.' '__')
out="$HOME/copia-${safe_name}-$(date +%Y%m%d-%H%M%S).json"

# O JSON vai só para o arquivo; mensagens e erros do comando ficam na tela. Se o
# comando falhar (e-mail sem conta, banco fora), nenhum arquivo é deixado.
if ! docker compose --env-file "$env_file" -f "$compose_file" --profile admin \
  run --rm -T account-export --email "$email" > "$out"; then
  rm -f "$out"
  fail "a cópia não foi gerada"
fi

chmod 600 "$out"
printf '\nCópia gerada: %s\n' "$out"
printf 'Próximos passos (o arquivo tem dados de saúde):\n'
printf '  1. Confirme que quem pediu é o dono da conta.\n'
printf '  2. Baixe:   scp <vps>:%s .\n' "$out"
printf '  3. Apague da VPS:   shred -u %s\n' "$out"
printf '  4. Compacte em .zip com senha e envie a senha por outro canal.\n'
