# Produção do Cadência

Esta estrutura implanta o Cadência de forma isolada na VPS Oracle:

```text
Internet
  -> Cloudflare Tunnel dedicado
  -> frontend (rede cadencia_edge)
  -> api (redes cadencia_edge e cadencia_data)
  -> PostgreSQL (rede cadencia_data, sem porta publica)
```

Nenhum serviço desta composição publica portas no host. O Cloudflare Tunnel é o único componente com saída para a internet. O PostgreSQL não recebe hostname, rota pública ou porta exposta.

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

Na raiz do repositório:

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

Antes de colocar dados reais, valide:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml ps
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml exec api wget -qO- http://127.0.0.1:8080/ready
```

## Atualização e migrações

Após revisar e atualizar o repositório, execute o `build`, aplique as migrações pelo perfil `maintenance` e só então reinicie os serviços de aplicação. O script registra cada arquivo SQL aplicado em `cadencia_schema_migrations`, portanto uma migração já concluída não é reaplicada.

## Backup e restauração

O script `scripts/backup-postgres.sh` cria um dump PostgreSQL no formato customizado, verifica sua leitura com `pg_restore --list` e conserva 14 dias por padrão. As unidades em `systemd/` programam essa rotina diariamente às 03:30 UTC, preservando a execução pendente depois de uma indisponibilidade da VPS.

Exemplo manual, na VPS:

```sh
sudo CADENCIA_BACKUP_DIR=/var/backups/cadencia \
  bash infrastructure/cadencia/scripts/backup-postgres.sh
```

Antes de automatizar o backup, o destino e a rotina de cópia externa devem ser definidos. A restauração será testada em banco/volume separado antes de qualquer restauração no banco de produção.

Na implantação, o diretório de backup será criado com acesso do usuário `ubuntu`, e as unidades serão copiadas para `/etc/systemd/system/`. Elas só serão ativadas depois da primeira restauração em ambiente isolado.

## Segurança operacional

- O proprietário do banco (`POSTGRES_USER`) serve apenas para operações administrativas: inicialização, migrações e backup.
- A API usa `POSTGRES_APP_USER`, sem superusuário, criação de banco ou criação de roles.
- A conexão entre API e PostgreSQL fica na rede Docker privada. Por isso, `sslmode=disable` é aceitável apenas dentro dessa rede local; não use essa configuração para uma conexão externa.
- O token do Cloudflare Tunnel deve ficar somente no `.env.production` da VPS.
