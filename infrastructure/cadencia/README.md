# Produção do Cadência

Esta estrutura implanta o Cadência de forma isolada na VPS Oracle. A composição inicial entrou em produção em 2 de setembro de 2026. Em 4 de setembro, a versão funcional `5fbc668` foi publicada com a migração `000015`, seguida do commit `c768ef7`, que atualizou a comunicação da versão `0.7.0` no frontend:

```text
Internet
  -> Cloudflare Tunnel dedicado
  -> frontend (rede cadencia_edge)
  -> api (redes cadencia_edge e cadencia_data)
  -> PostgreSQL (rede cadencia_data, sem porta publica)
```

Nenhum serviço desta composição publica portas no host. O Cloudflare Tunnel é o único componente que encaminha tráfego público para o Cadência. O PostgreSQL não recebe hostname, rota pública ou porta exposta.

O serviço `tunnel` usa o digest fixo `cloudflare/cloudflared@sha256:e39ee8da81ad5e05d77f38d2f51c60ca51bf2a8450ac3abab50c17fdb91d91bf`, correspondente ao binário `cloudflared 2026.7.3` em `linux/arm64` validado na VPS. Esse pin foi publicado no commit `61d7939` e aplicado no deploy de 11 de setembro de 2026.

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

### Cópia completa dos dados de um atleta (LGPD)

A planilha em Configurações traz só o essencial e deixa de fora os dados de saúde. Quando um atleta pedir a cópia completa (a política promete resposta em até 15 dias), gere o arquivo na VPS com o comando administrativo. Ele só lê o banco, não abre portas, não tem rede de saída e não é acessível pela API:

```sh
sh infrastructure/cadencia/scripts/account-export.sh atleta@exemplo.com
```

O script grava `~/copia-<e-mail>-<data>.json` (legível só pelo seu usuário), mostra os próximos passos e não deixa arquivo se algo falhar. Por baixo, ele roda o serviço `account-export` do perfil `admin` do Compose. O arquivo traz conta, perfil, metas, disponibilidade, limitações, check-ins, avaliações, planos, treinos, sessões, feedbacks, atividades importadas, aceites dos termos e mensagens de feedback. Não traz senha nem códigos de sessão. O log (na tela, não no arquivo) registra o e-mail consultado e a hora, sem o conteúdo. Um e-mail sem conta, ou inválido, termina com erro e sem arquivo.

Cuidados, porque o arquivo tem dados de saúde:

1. **Confirme quem pediu.** Responda ao e-mail cadastrado ou confirme a identidade de outra forma antes de enviar qualquer coisa.
2. **Baixe e apague.** Copie o arquivo para o seu computador (`scp`) e apague-o da VPS com `shred -u`; o script mostra os dois comandos já com o nome do arquivo.
3. **Envie protegido.** Compacte em `.zip` com senha forte e mande a senha por outro canal, como o WhatsApp. Não anexe o JSON puro.
4. **Anote o pedido** (data, quem pediu, quando respondeu) onde você controla os pedidos de privacidade.

Se o app crescer, o próximo passo é um fluxo automatizado com verificação de identidade; hoje, com poucos atletas, o procedimento manual é suficiente.

O esquema versionado vai até `000035` e a produção está sincronizada com ele (deploy de 1 de outubro de 2026, commit `40898f8`). A versão visível ao usuário é `0.37.0`. O histórico de cada deploy fica em [`docs/STATUS.md`](../../docs/STATUS.md) e em [`docs/changelog/project-status-history.md`](../../docs/changelog/project-status-history.md).

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

## Deploy automatizado, monitoramento e cópia externa

O script `scripts/deploy.sh` executa o procedimento completo e para no primeiro erro: fast-forward do checkout, backup verificado, build, listagem das migrações pendentes (`DRY_RUN=1`), aplicação em ordem, recriação de `api`, `frontend` e `tunnel`, espera por `/ready` e smoke test público e autenticado.

```sh
sh infrastructure/cadencia/scripts/deploy.sh          # pergunta antes de aplicar
sh infrastructure/cadencia/scripts/deploy.sh --yes    # sem confirmação
```

Para o smoke test autenticado, crie uma conta de teste dedicada com e-mail confirmado e exporte `CADENCIA_SMOKE_EMAIL` e `CADENCIA_SMOKE_PASSWORD` (por exemplo em `/etc/cadencia/smoke.env`, lido antes de chamar o script). Sem elas, essa etapa é pulada com um aviso: healthchecks sozinhos não comprovam que o schema atende ao código.

Compatibilidade de schema: a API conhece a lista de migrações que exige (`backend/internal/database/schema.go`). Em produção ela se recusa a iniciar, e `/ready` responde `503 schema_behind`, enquanto faltar alguma. Por isso o healthcheck do container usa `/ready` e o tunnel só sobe com a API pronta. Ao criar uma migração, atualize essa lista (o teste `TestRequiredMigrationsMatchFiles` avisa se esquecer).

O `migrate.sh` também aceita `DRY_RUN=1` (apenas lista), `MIGRATE_STRICT=1` (falha se uma migração já aplicada foi modificada) e registra o checksum de cada arquivo aplicado.

Configuração opcional do backup, em `/etc/cadencia/backup.env` (lido pelo `cadencia-backup.service`):

```sh
# Cópia externa (rclone). Use um remote "crypt": o dump contém dados pessoais e de saúde.
CADENCIA_OFFSITE_REMOTE=cadencia-crypt:
# A VM não tem permissão de apagar no bucket; a retenção externa é uma regra de
# ciclo de vida do próprio bucket. Use =1 só se a VM puder apagar objetos.
# CADENCIA_OFFSITE_PRUNE=1
# Monitor de ping (Healthchecks.io ou similar): recebe /start, sucesso e /fail.
CADENCIA_HEALTHCHECK_URL=https://hc-ping.com/<uuid>
```

O `test-restore.sh` restaura o dump mais recente em um PostgreSQL descartável e confere o registro de migrações; o timer `cadencia-restore-test.timer` o executa todo dia 1 às 05:00 UTC (instale as duas unidades como as demais e use `/etc/cadencia/monitoring.env` para o ping). Um backup só é considerado bom depois de restaurado.

Os serviços têm limites de memória (`API_MEM_LIMIT`, `FRONTEND_MEM_LIMIT`, `POSTGRES_MEM_LIMIT`) e rotação de logs (10 MB × 3). As imagens são fixadas por digest; o Dependabot propõe as atualizações.

## Atualização e migrações

Após revisar e atualizar o repositório, crie um backup, execute o `build`, aplique as migrações pelo perfil `maintenance` e só então reinicie os serviços de aplicação. O script registra cada arquivo SQL aplicado em `cadencia_schema_migrations`, portanto uma migração já concluída não é reaplicada. Para a atualização de dependências do commit `41638da`, não houve mudança de esquema e somente a API foi reconstruída.

Para uma alteração somente de interface, como o ajuste dos gráficos mobile do commit `33de28a`, o procedimento usado em 3 de setembro de 2026 foi: atualizar o checkout por fast-forward, criar o backup preventivo, reconstruir somente `frontend`, executar `up -d frontend` e validar o domínio oficial. API, PostgreSQL, túnel e os demais aplicativos da VPS não precisam ser recriados quando não há mudança correspondente.

No deploy funcional de 4 de setembro de 2026, o commit `5fbc668` foi atualizado por fast-forward, o backup `cadencia-20260905T003553Z.dump` foi criado e verificado, a migração `000015` foi aplicada pelo perfil `maintenance` e as imagens de API e frontend foram reconstruídas. Os containers `api` e `frontend` foram recriados; PostgreSQL e túnel permaneceram ativos. Em seguida, o commit `c768ef7` atualizou somente o frontend para publicar a nota da versão `0.7.0`.

No deploy de 5 de setembro de 2026, o commit `9d8c624` foi atualizado por fast-forward, o backup `cadencia-20260905T221541Z.dump` foi criado e verificado, a migração `000016` foi aplicada pelo perfil `maintenance` e as imagens da API e do frontend foram reconstruídas. API e frontend ficaram saudáveis; PostgreSQL e túnel permaneceram ativos. A versão `0.8.0` e a release `v0.8.0` publicam o piloto de intervalos aeróbicos XCO.

No deploy de 11 de setembro de 2026, o commit `61d7939` foi atualizado por fast-forward. O backup preventivo `cadencia-20260911T234851Z.dump` foi criado e verificado; não havia migração nova para aplicar. As imagens da API e do frontend foram reconstruídas e os serviços de aplicação e o Tunnel foram recriados; PostgreSQL permaneceu ativo e saudável. A API interna respondeu `{"status":"ready"}`, os dois domínios públicos retornaram HTTP 200 e a release `v0.12.0` foi publicada no GitHub.

No deploy de 13 de setembro de 2026, o commit `84b653b` foi atualizado por fast-forward. O backup preventivo `cadencia-20260913T161932Z.dump` foi criado e verificado; as migrações `000020` e `000021` foram aplicadas pelo perfil `maintenance`. As imagens da API e do frontend foram reconstruídas e os dois serviços foram recriados; PostgreSQL e Tunnel permaneceram ativos. `/health` e `/ready` responderam corretamente, os dois domínios públicos retornaram HTTP 200 e a versão `0.20.0` foi confirmada no HTML público. A release [v0.20.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.20.0) foi publicada no GitHub.

Em 13 de setembro de 2026, o commit `c5d8822` foi atualizado por fast-forward. O backup preventivo `cadencia-20260913T190527Z.dump` foi criado e verificado; não houve migração nova. As imagens da API e do frontend foram reconstruídas e os dois serviços foram recriados; PostgreSQL e Tunnel permaneceram ativos. `/ready` respondeu `{"status":"ready"}`, os dois domínios públicos retornaram HTTP 200 e os quatro serviços ficaram saudáveis. A versão visível permaneceu `0.20.0`; não foi criada nova release.

Em 13 de setembro de 2026, o commit `f0fec8b` foi atualizado por fast-forward. O backup preventivo `cadencia-20260913T202532Z.dump` foi criado e verificado; não houve migração nova. As imagens da API e do frontend foram reconstruídas e os dois serviços foram recriados; PostgreSQL e Tunnel permaneceram ativos. `/ready` respondeu `{"status":"ready"}`, os dois domínios públicos retornaram HTTP 200 e os quatro serviços ficaram saudáveis. A versão visível permaneceu `0.20.0`; não foi criada nova release.

Em 14 de setembro de 2026, o commit funcional `f2f8192` foi atualizado por fast-forward. O backup preventivo `cadencia-20260914T224807Z.dump` foi criado e verificado; as migrações `000024` a `000029` foram aplicadas em ordem pelo perfil `maintenance`. As imagens da API e do frontend foram reconstruídas e os serviços de aplicação foram recriados; PostgreSQL e Tunnel permaneceram ativos. `/health` e `/ready` responderam corretamente, os dois domínios oficiais retornaram HTTP 200 e os quatro serviços ficaram saudáveis. A versão visível `0.30.0` foi confirmada no HTML público. A leitura autenticada de `GET /v1/plans/current` também foi validada, incluindo a tela de novidades e a explicação do plano.

Na mesma data, uma validação autenticada intermediária revelou que `GET /v1/plans/current` retornava `500`, embora os healthchecks estivessem verdes. A API consultava campos ainda não presentes no banco; a checagem SQL confirmou a defasagem. Foi criado e validado o backup `cadencia-20260914T112526Z.dump`, e o perfil `maintenance` aplicou em ordem as migrações `000022_limitation_context` e `000023_feedback_context`. Depois, antes do fechamento da versão `0.30.0`, as migrações `000024` a `000029` foram igualmente aplicadas e o plano foi validado novamente. Regra permanente: healthchecks devem ser acompanhados da conferência de `cadencia_schema_migrations` e de uma leitura autenticada de uma rota crítica antes de encerrar o deploy.

No deploy de 15 de setembro de 2026, o commit `95f2c27` foi atualizado por fast-forward. O backup preventivo `cadencia-20260915T113044Z.dump` foi criado e verificado; a migração `000030_profile_safety_context` foi aplicada pelo perfil `maintenance`. As imagens da API e do frontend foram reconstruídas, e API, frontend e Tunnel foram recriados; PostgreSQL permaneceu ativo e saudável. `/health` e `/ready` responderam corretamente, os dois domínios públicos retornaram HTTP 200 e a versão `0.32.0` foi confirmada no HTML público. A leitura autenticada de `/v1/plans/current` ainda precisa ser conferida manualmente nesta publicação porque o controlador de navegador não ficou disponível durante a execução.

O deploy oficial deve sempre terminar em `https://cadencia.devsaulo.com.br` e `https://cadencia-api.devsaulo.com.br`, pela composição Docker desta pasta e pelo Cloudflare Tunnel dedicado. O ambiente Sites não faz parte da produção do Cadência e não deve ser usado como destino alternativo.

## Backup e restauração

O script `scripts/backup-postgres.sh` cria um dump PostgreSQL no formato customizado, verifica sua leitura com `pg_restore --list` e conserva 14 dias por padrão. A unidade `cadencia-backup.timer` está habilitada na VPS e executa essa rotina diariamente às 03:30 UTC, preservando a execução pendente depois de uma indisponibilidade da VPS.

Exemplo manual, na VPS:

```sh
sudo CADENCIA_BACKUP_DIR=/var/backups/cadencia \
  bash infrastructure/cadencia/scripts/backup-postgres.sh
```

O diretório de produção é `/var/backups/cadencia`, com acesso do usuário `ubuntu`. O backup preventivo mais recente é `cadencia-20260915T113044Z.dump` (UTC), criado antes da aplicação da migração `000030`; a validação estrutural do arquivo ocorreu automaticamente. Os backups anteriores `cadencia-20260914T224807Z.dump`, `cadencia-20260914T112526Z.dump` e `cadencia-20260902T104801Z.dump` também permanecem registrados e validados. O teste completo de restauração do segundo foi concluído em um PostgreSQL 17 temporário, sem alterar a produção. A cópia externa dos dumps ainda está pendente.

## Segurança operacional

- O proprietário do banco (`POSTGRES_USER`) serve apenas para operações administrativas: inicialização, migrações e backup.
- A API usa `POSTGRES_APP_USER`, sem superusuário, criação de banco ou criação de roles.
- A conexão entre API e PostgreSQL fica na rede Docker privada. Por isso, `sslmode=disable` é aceitável apenas dentro dessa rede local; não use essa configuração para uma conexão externa.
- O token do Cloudflare Tunnel deve ficar somente no `.env.production` da VPS.
- A API usa Go 1.26 na imagem de build (`infrastructure/cadencia/Dockerfile.api`).
- A IA explicativa permanece desligada por padrão (`AI_ENABLED=false`). Na VPS, ela está temporariamente ativa com `AI_PROVIDER=worker`, usando a rota protegida do Cloudflare; o serviço Ollama está instalado, mas parado após o teste de capacidade. O compose encaminha os limites e, opcionalmente, `AI_WORKER_URL`/`AI_WORKER_TOKEN` para a API, sem publicar a porta 11434. Reavalie a ativação local somente após uma nova medição de capacidade.
- O Worker Cloudflare possui a rota protegida `/cadencia/explanation`, separada do endpoint legado usado por outros projetos. A versão ativa usa `openai/gpt-oss-20b` com `max_completion_tokens: 512` e `reasoning_effort: 'low'`, rejeitando respostas sem `finish_reason: 'stop'` para acionar o fallback determinístico. O segredo `CADENCIA_WORKER_TOKEN` está configurado no Worker e o valor correspondente fica somente no `.env.production` do Cadência; uma chamada sintética e testes autenticados foram validados. Nunca coloque esse token no frontend ou no repositório.
- O PostgreSQL não deve receber porta publicada, hostname público ou regra no Cloudflare.
- O hardening das portas dos demais aplicativos da VPS é uma atividade separada; não altere seus containers por este compose.

Na auditoria de 2 de setembro de 2026, os serviços externos continuavam fora desta composição. O Nginx Proxy Manager usa Tailscale para `casaos.oraclecloud.com.br` (`100.67.151.30:8888`) e `immich.photo.com.br` (`100.67.151.30:2283`). A porta `8123` é do Home Assistant e a `8888` é do `casaos-gateway`; existe ainda um `cloudflared-tunnel` separado para outros aplicativos. A porta pública `2283` foi bloqueada na cadeia `DOCKER-USER` somente pela interface `enp0s6`, preservando Tailscale, loopback e o funcionamento do Immich. Novas conexões SSH também foram bloqueadas em `enp0s6`, mantendo o acesso administrativo validado pelo IP Tailscale `100.67.151.30`. As regras TCP públicas `22`, `81`, `2283`, `8096` e `8097` foram removidas da Oracle Cloud; restaram somente ICMP.

### Auditoria de segurança da VPS (01/10/2026)

Auditoria somente de leitura, feita por SSH e por testes de fora. Escopo: o servidor (host) e o Cadência. Os demais aplicativos da VPS não foram alterados.

| Item | Situação |
| --- | --- |
| SSH | Só por chave (`PasswordAuthentication no`); conexões novas na porta 22 pela interface pública são descartadas por regra do firewall e o acesso é pela Tailscale. 0 tentativas falhas nas últimas 24 horas. |
| Exposição pública | Testadas 18 portas no IP público (22, 80, 443, 81, 111, 139, 445, 3000, 3001, 3011, 4443, 5678, 8080, 8081, 8082, 8091, 8181, 9000): todas fechadas. O firewall da nuvem bloqueia tudo; o Cadência entra só pelo Cloudflare Tunnel e não publica porta nenhuma. |
| Segredos | `.env.production` em 600; `/etc/cadencia/*.env` em 640 (root:ubuntu); backups em 600, com cópia externa criptografada e regra de exclusão em 60 dias. |
| Atualizações | `unattended-upgrades` ativo. Havia 46 pacotes atualizáveis e um reinício pendente (kernel). |
| Contêineres do Cadência | Rodam sem root (API, frontend e túnel), com limite de memória, logs rotacionados e imagens fixadas por hash; o PostgreSQL fica numa rede interna, sem porta publicada. |
| Lacuna encontrada | Os contêineres mantinham as capacidades padrão do kernel, podiam elevar privilégios e tinham o sistema de arquivos gravável. |

**Mudança feita (`compose.production.yaml`):** API, frontend, túnel e o comando `account-export` passam a rodar sem nenhuma capacidade (`cap_drop: ALL`), com `no-new-privileges`, sistema de arquivos somente leitura (só `/tmp` gravável, em memória, `noexec`) e `pids_limit: 256`. Testado localmente com as imagens de produção: a API e o frontend sobem e ficam saudáveis; cadastro, geração de plano, importação de arquivo e planilha funcionam sem erros de escrita; o túnel inicia normalmente; a cópia completa chega ao banco. O PostgreSQL ficou de fora (precisa gravar no volume e trocar de usuário ao iniciar). Se algum serviço falhar depois do deploy, o primeiro passo é remover a linha `<<: *hardening` do serviço e recriá-lo.

**Pendente, com decisão do dono do produto:**

1. **Reinício da VPS: feito em 01/10/2026** (kernel `6.17` → `7.0.0-1012-oracle`, sem reinício pendente). Voltou em cerca de 50 segundos; os 19 contêineres, o Samba, o CasaOS, o Tailscale, o SSH e os timers do Cadência voltaram sozinhos, sem contêiner doente, e a API respondeu `/ready` e o frontend 200.
2. **Reduzir o SSH** (opcional, baixo ganho porque a porta não é pública): `X11Forwarding no` e `MaxAuthTries 3` em um arquivo de `/etc/ssh/sshd_config.d/`, validando com `sshd -t` antes de recarregar, sem fechar a sessão atual.
3. **`fail2ban` desligado:** sem ganho enquanto a porta 22 estiver descartada na interface pública; só vale se ela for reaberta.
4. **Samba (139/445) e rpcbind (111)** escutam em todas as interfaces do host. O dono do produto usa o Samba, então fica como está. Não estão acessíveis da internet (testado), apenas na rede local e na Tailscale. Se quiser reduzir, dá para limitar o Samba à interface da Tailscale/rede local em `smb.conf` (`interfaces`/`bind interfaces only`); o rpcbind só serve a NFS e pode ser desligado se não houver NFS.
5. **Firewall do host** com política `ACCEPT` por padrão: a proteção pública depende do firewall da Oracle. Trocar para política de bloqueio exige regras para Docker e Tailscale e pode derrubar os outros aplicativos; fica como melhoria futura, com janela de teste.
6. **Cloudflare, conferido no painel em 01/10/2026:** modo SSL "Completo" (automático; não trocar para "Full (strict)" sem antes olhar os outros registros DNS da zona, porque o modo vale para ela inteira e o Cadência chega à VPS por túnel, que já é criptografado); certificado universal gerenciado, válido até 26/11/2026; HSTS ativo (6 meses, com subdomínios, sem pré-carga); TLS 1.3 ligado; versão mínima do TLS **subida de 1.0 para 1.2** (nas 24 horas anteriores só houve tráfego em 1.2 e 1.3). Confirmado de fora: TLS 1.0 e 1.1 recusados com alerta "protocol version" pela Cloudflare, 1.2 e 1.3 aceitos, e HTTP redireciona para HTTPS nos dois domínios. Limite de requisições criado e testado: regra "Limitar autenticação" (host `cadencia-api.devsaulo.com.br` e caminho `/v1/auth/`; 10 requisições por 10 segundos por IP; bloqueio de 10 segundos, o mínimo do plano gratuito, que só permite uma regra); numa rajada de 25 requisições, as 10 primeiras passaram e as 15 seguintes receberam 429, `/health` não foi afetado e o acesso voltou após o bloqueio. Já existia a regra "Permitir UpTime Kuma" (ignora o IP da própria VPS). Monitoramento de transparência de certificado ativado pelo dono do produto (e-mail quando alguém emitir certificado para o domínio). **Bot Fight Mode deixado desligado de propósito:** no plano gratuito não aceita exceções por regra e pode desafiar tráfego automatizado legítimo (testes de fumaça do deploy, Uptime Kuma, n8n e, no futuro, os webhooks do Strava, que precisam responder 200 em até 2 segundos); o ganho é pequeno porque a autenticação já tem limite no código e na Cloudflare. Reavaliar se aparecerem acessos suspeitos. Painel da Cloudflare concluído.
