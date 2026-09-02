# Produção do Cadência

Esta estrutura implanta o Cadência de forma isolada na VPS Oracle. Ela está em produção desde 2 de setembro de 2026:

```text
Internet
  -> Cloudflare Tunnel dedicado
  -> frontend (rede cadencia_edge)
  -> api (redes cadencia_edge e cadencia_data)
  -> PostgreSQL (rede cadencia_data, sem porta publica)
```

Nenhum serviço desta composição publica portas no host. O Cloudflare Tunnel é o único componente que encaminha tráfego público para o Cadência. O PostgreSQL não recebe hostname, rota pública ou porta exposta.

URLs em produção:

- `https://cadencia.devsaulo.com.br` -> `frontend:3000`
- `https://cadencia-api.devsaulo.com.br` -> `api:8080`

Containers atuais: `cadencia-api-1`, `cadencia-frontend-1`, `cadencia-postgres-1` e `cadencia-tunnel-1`.

## Pré-requisitos

- Docker e Docker Compose Plugin na VPS;
- repositório clonado em `/home/ubuntu/apps/cadencia`;
- túnel Cloudflare **dedicado** criado no painel;
- hostnames públicos configurados no túnel:
  - `cadencia.devsaulo.com.br` -> `http://frontend:3000`
  - `cadencia-api.devsaulo.com.br` -> `http://api:8080`

Os nomes `frontend` e `api` são resolvidos somente dentro da rede Docker `cadencia_edge`.

## Preparação inicial

1. Copie `.env.production.example` para `.env.production` dentro desta pasta.
2. Gere senhas fortes para `POSTGRES_PASSWORD` e `POSTGRES_APP_PASSWORD`; use apenas letras, números, hífen e sublinhado.
3. Atualize `DATABASE_URL` com a mesma senha de `POSTGRES_APP_PASSWORD`.
4. Cole o token do túnel dedicado em `CLOUDFLARE_TUNNEL_TOKEN`.
5. Nunca envie `.env.production`, backups ou tokens ao Git.

## Primeiro deploy

Na raiz do repositório. O procedimento abaixo é reexecutável para uma instalação nova ou uma atualização controlada:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml build
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml up -d postgres
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml --profile maintenance run --rm migrate
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml up -d api frontend tunnel
```

Após o deploy, valide:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml ps
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml exec api wget -qO- http://127.0.0.1:8080/ready
```

## Atualização e migrações

Após revisar e atualizar o repositório, crie um backup, execute o `build`, aplique as migrações pelo perfil `maintenance` e só então reinicie os serviços de aplicação. O script registra cada arquivo SQL aplicado em `cadencia_schema_migrations`, portanto uma migração já concluída não é reaplicada. Para a atualização de dependências do commit `41638da`, não houve mudança de esquema e somente a API foi reconstruída.

## Backup e restauração

O script `scripts/backup-postgres.sh` cria um dump PostgreSQL no formato customizado, verifica sua leitura com `pg_restore --list` e conserva 14 dias por padrão. A unidade `cadencia-backup.timer` está habilitada na VPS e executa essa rotina diariamente às 03:30 UTC, preservando a execução pendente depois de uma indisponibilidade da VPS.

Exemplo manual, na VPS:

```sh
sudo CADENCIA_BACKUP_DIR=/var/backups/cadencia \
  bash infrastructure/cadencia/scripts/backup-postgres.sh
```

O diretório de produção é `/var/backups/cadencia`, com acesso do usuário `ubuntu`. O backup preventivo do deploy mais recente foi `cadencia-20260902T104801Z.dump`. A validação estrutural do arquivo já ocorre automaticamente; o teste completo de restauração em banco/volume separado e a cópia externa dos dumps ainda estão pendentes.

## Segurança operacional

- O proprietário do banco (`POSTGRES_USER`) serve apenas para operações administrativas: inicialização, migrações e backup.
- A API usa `POSTGRES_APP_USER`, sem superusuário, criação de banco ou criação de roles.
- A conexão entre API e PostgreSQL fica na rede Docker privada. Por isso, `sslmode=disable` é aceitável apenas dentro dessa rede local; não use essa configuração para uma conexão externa.
- O token do Cloudflare Tunnel deve ficar somente no `.env.production` da VPS.
- A API usa Go 1.25 na imagem de build (`infrastructure/cadencia/Dockerfile.api`).
- O PostgreSQL não deve receber porta publicada, hostname público ou regra no Cloudflare.
- O hardening das portas dos demais aplicativos da VPS é uma atividade separada; não altere seus containers por este compose.
