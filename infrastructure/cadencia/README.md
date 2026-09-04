# Produção do Cadência

Esta estrutura implanta o Cadência de forma isolada na VPS Oracle. A composição inicial entrou em produção em 2 de setembro de 2026 e recebeu o último ajuste de interface no commit `33de28a` em 3 de setembro:

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

- Docker e Docker Compose Plugin na VPS, com suporte a `gw_priority` (Compose 2.33.1 ou superior);
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

Para receber o resumo semanal de feedback, preencha `FEEDBACK_DIGEST_TO` com um único endereço administrativo. Deixe a variável vazia para manter o recurso desativado. O endereço não é exibido aos atletas e não é usado pelo fluxo de feedback.

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

### Resumo semanal de feedback

O resumo roda em um comando curto, fora do processo HTTP. Depois de aplicar as migrações e reconstruir a imagem, instale as duas unidades systemd no host e habilite o timer:

```sh
sudo cp infrastructure/cadencia/systemd/cadencia-feedback-digest.service /etc/systemd/system/
sudo cp infrastructure/cadencia/systemd/cadencia-feedback-digest.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now cadencia-feedback-digest.timer
sudo systemctl list-timers cadencia-feedback-digest.timer
```

O timer executa às segundas-feiras, às 11:00 UTC (08:00 no horário de São Paulo), e pode recuperar uma execução perdida por causa de `Persistent=true`. O job usa o perfil Compose `digest`, consulta o PostgreSQL pela rede interna e acessa o Resend por uma rede de saída dedicada; envia no máximo 50 relatos pendentes em uma única mensagem e marca cada relato depois do envio bem-sucedido. Ele não fica residente e não publica portas. Para testar manualmente na VPS:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml --profile digest run --rm feedback-digest
```

O primeiro deploy que incluir o recurso deve executar as migrações `000013_user_feedback` e `000014_feedback_digest` pelo perfil `maintenance` antes de habilitar o timer. Se `FEEDBACK_DIGEST_TO` estiver vazio, o comando encerra sem enviar e-mail.

No checkout atual, a migração `000015` registra as fontes científicas do catálogo e ainda está pendente na produção, que permanece até `000014`. Qualquer deploy que inclua o catálogo deve passar por revisão, backup verificável, execução ordenada pelo perfil `maintenance`, validação e autorização explícita; a atualização também deve registrar a novidade em `frontend/lib/release.ts` antes da publicação.

O Ollama é opcional e não é iniciado pelo comando acima. Ele foi instalado na VPS e permanece parado após o teste de capacidade; a produção usa temporariamente o Worker remoto para evitar sobrecarga. O padrão seguro continua sendo `AI_ENABLED=false`. Para preparar o serviço somente na rede interna do Cadência:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml up -d ollama
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml exec ollama \
  ollama pull qwen3:4b-instruct
```

O serviço tem limite de 4 GiB de memória, uma execução simultânea e nenhum `ports:` publicado; `11434` fica acessível somente pela rede Docker privada. O modelo `qwen3:4b-instruct` já foi baixado e testado, mas uma chamada levou cerca de 72 segundos e consumiu praticamente 100% do limite de CPU. Mantenha o serviço parado até existir folga ou otimização suficiente; a produção usa o Worker remoto durante esta fase.

Após o deploy, valide:

```sh
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml ps
docker compose --env-file infrastructure/cadencia/.env.production \
  -f infrastructure/cadencia/compose.production.yaml exec api wget -qO- http://127.0.0.1:8080/ready
```

## Atualização e migrações

Após revisar e atualizar o repositório, crie um backup, execute o `build`, aplique as migrações pelo perfil `maintenance` e só então reinicie os serviços de aplicação. O script registra cada arquivo SQL aplicado em `cadencia_schema_migrations`, portanto uma migração já concluída não é reaplicada. Para a atualização de dependências do commit `41638da`, não houve mudança de esquema e somente a API foi reconstruída.

Para uma alteração somente de interface, como o ajuste dos gráficos mobile do commit `33de28a`, o procedimento usado em 3 de setembro de 2026 foi: atualizar o checkout por fast-forward, criar o backup preventivo, reconstruir somente `frontend`, executar `up -d frontend` e validar o domínio oficial. API, PostgreSQL, túnel e os demais aplicativos da VPS não precisam ser recriados quando não há mudança correspondente.

O deploy oficial deve sempre terminar em `https://cadencia.devsaulo.com.br` e `https://cadencia-api.devsaulo.com.br`, pela composição Docker desta pasta e pelo Cloudflare Tunnel dedicado. O ambiente Sites não faz parte da produção do Cadência e não deve ser usado como destino alternativo.

## Backup e restauração

O script `scripts/backup-postgres.sh` cria um dump PostgreSQL no formato customizado, verifica sua leitura com `pg_restore --list` e conserva 14 dias por padrão. A unidade `cadencia-backup.timer` está habilitada na VPS e executa essa rotina diariamente às 03:30 UTC, preservando a execução pendente depois de uma indisponibilidade da VPS.

Exemplo manual, na VPS:

```sh
sudo CADENCIA_BACKUP_DIR=/var/backups/cadencia \
  bash infrastructure/cadencia/scripts/backup-postgres.sh
```

O diretório de produção é `/var/backups/cadencia`, com acesso do usuário `ubuntu`. O backup preventivo do deploy mais recente é `cadencia-20260904T023630Z.dump` (UTC). A validação estrutural do arquivo ocorre automaticamente. O backup anterior `cadencia-20260902T104801Z.dump` também foi restaurado com sucesso em um PostgreSQL 17 temporário: foram confirmadas 15 tabelas públicas e `cadencia_schema_migrations`, e o ambiente temporário foi removido sem alterar a produção. A cópia externa dos dumps ainda está pendente.

## Segurança operacional

- O proprietário do banco (`POSTGRES_USER`) serve apenas para operações administrativas: inicialização, migrações e backup.
- A API usa `POSTGRES_APP_USER`, sem superusuário, criação de banco ou criação de roles.
- A conexão entre API e PostgreSQL fica na rede Docker privada. Por isso, `sslmode=disable` é aceitável apenas dentro dessa rede local; não use essa configuração para uma conexão externa.
- O token do Cloudflare Tunnel deve ficar somente no `.env.production` da VPS.
- A API usa Go 1.25 na imagem de build (`infrastructure/cadencia/Dockerfile.api`).
- A IA explicativa permanece desligada por padrão (`AI_ENABLED=false`). Na VPS, ela está temporariamente ativa com `AI_PROVIDER=worker`, usando a rota protegida do Cloudflare; o serviço Ollama está instalado, mas parado após o teste de capacidade. O compose encaminha os limites e, opcionalmente, `AI_WORKER_URL`/`AI_WORKER_TOKEN` para a API, sem publicar a porta 11434. Reavalie a ativação local somente após uma nova medição de capacidade.
- O Worker Cloudflare possui a rota protegida `/cadencia/explanation`, separada do endpoint legado usado por outros projetos. A versão ativa usa `openai/gpt-oss-20b` com `max_completion_tokens: 512` e `reasoning_effort: 'low'`, rejeitando respostas sem `finish_reason: 'stop'` para acionar o fallback determinístico. O segredo `CADENCIA_WORKER_TOKEN` está configurado no Worker e o valor correspondente fica somente no `.env.production` do Cadência; uma chamada sintética e testes autenticados foram validados. Nunca coloque esse token no frontend ou no repositório.
- O PostgreSQL não deve receber porta publicada, hostname público ou regra no Cloudflare.
- O hardening das portas dos demais aplicativos da VPS é uma atividade separada; não altere seus containers por este compose.

Na auditoria de 2 de setembro de 2026, os serviços externos continuavam fora desta composição. O Nginx Proxy Manager usa Tailscale para `casaos.oraclecloud.com.br` (`100.67.151.30:8888`) e `immich.photo.com.br` (`100.67.151.30:2283`). A porta `8123` é do Home Assistant e a `8888` é do `casaos-gateway`; existe ainda um `cloudflared-tunnel` separado para outros aplicativos. A porta pública `2283` foi bloqueada na cadeia `DOCKER-USER` somente pela interface `enp0s6`, preservando Tailscale, loopback e o funcionamento do Immich. Novas conexões SSH também foram bloqueadas em `enp0s6`, mantendo o acesso administrativo validado pelo IP Tailscale `100.67.151.30`. As regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud; restaram somente ICMP.
