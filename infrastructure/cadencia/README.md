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

O diretório de produção é `/var/backups/cadencia`, com acesso do usuário `ubuntu`. O backup preventivo do deploy mais recente foi `cadencia-20260902T104801Z.dump`. A validação estrutural do arquivo ocorre automaticamente. Em 2 de setembro de 2026, esse dump também foi restaurado com sucesso em um PostgreSQL 17 temporário: foram confirmadas 15 tabelas públicas e `cadencia_schema_migrations`, e o ambiente temporário foi removido sem alterar a produção. A cópia externa dos dumps ainda está pendente.

## Segurança operacional

- O proprietário do banco (`POSTGRES_USER`) serve apenas para operações administrativas: inicialização, migrações e backup.
- A API usa `POSTGRES_APP_USER`, sem superusuário, criação de banco ou criação de roles.
- A conexão entre API e PostgreSQL fica na rede Docker privada. Por isso, `sslmode=disable` é aceitável apenas dentro dessa rede local; não use essa configuração para uma conexão externa.
- O token do Cloudflare Tunnel deve ficar somente no `.env.production` da VPS.
- A API usa Go 1.25 na imagem de build (`infrastructure/cadencia/Dockerfile.api`).
- A IA explicativa permanece desligada por padrão (`AI_ENABLED=false`). O compose encaminha os limites e, opcionalmente, `AI_WORKER_URL`/`AI_WORKER_TOKEN` para a API; o serviço Ollama ainda não é iniciado. Só ative depois de medir a VPS e configurar o Ollama em uma rede Docker interna, sem publicar a porta 11434.
- O Worker Cloudflare possui a rota protegida `/cadencia/explanation`, separada do endpoint legado usado por outros projetos. O segredo `CADENCIA_WORKER_TOKEN` está configurado no Worker e o valor correspondente fica somente no `.env.production` do Cadência; uma chamada sintética autenticada foi validada com o modelo `openai/gpt-oss-20b`. Nunca coloque esse token no frontend ou no repositório.
- O PostgreSQL não deve receber porta publicada, hostname público ou regra no Cloudflare.
- O hardening das portas dos demais aplicativos da VPS é uma atividade separada; não altere seus containers por este compose.

Na auditoria de 2 de setembro de 2026, os serviços externos continuavam fora desta composição. O Nginx Proxy Manager usa Tailscale para `casaos.oraclecloud.com.br` (`100.67.151.30:8888`) e `immich.photo.com.br` (`100.67.151.30:2283`). A porta `8123` é do Home Assistant e a `8888` é do `casaos-gateway`; existe ainda um `cloudflared-tunnel` separado para outros aplicativos. A porta pública `2283` foi bloqueada na cadeia `DOCKER-USER` somente pela interface `enp0s6`, preservando Tailscale, loopback e o funcionamento do Immich. Novas conexões SSH também foram bloqueadas em `enp0s6`, mantendo o acesso administrativo validado pelo IP Tailscale `100.67.151.30`. As regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud; restaram somente ICMP.
