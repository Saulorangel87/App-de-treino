# Cadência

Aplicação de planejamento adaptativo de treinos de ciclismo.

Versão atual: `0.6.0` — resumo semanal de feedback, PWA e identidade visual Cadência. Consulte o [release v0.6.0](https://github.com/Saulorangel87/App-de-treino/releases/tag/v0.6.0).

## Estrutura

- `frontend/`: aplicação React/TypeScript publicada na VPS Oracle por Docker e Cloudflare Tunnel.
- `backend/`: API REST em Go.
- `database/`: migrações PostgreSQL versionadas.
- `api/`: contrato OpenAPI.
- `infrastructure/`: configuração versionada para a VPS Oracle.
- `docs/`: decisões, regras, ciclo de vida e catálogo científico do produto.
- `docs/README.md`: índice da documentação e regra de atualização.
- `melhorias.md`: roadmap atual da próxima fase; `planejamento.md` preserva a visão e o histórico do projeto.

## Ambiente local

1. Copie `.env.example` para `.env` e use somente credenciais locais.
2. Inicie o PostgreSQL com `docker compose up -d postgres`.
3. Aplique os arquivos `database/migrations/*.up.sql` ainda pendentes, em ordem numérica. O esquema versionado inclui a migração `000015` (fontes científicas do catálogo); ela ainda precisa ser aplicada nos ambientes que estiverem em `000014`.
4. Execute a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
5. Execute o frontend a partir de `frontend/` com `npm run dev`.

O frontend nunca se conecta diretamente ao PostgreSQL. Todo acesso passa pela API Go.
A configuração local deste projeto usa a porta `5433` no `.env`, pois a `5432` já estava ocupada no Windows. O `.env.example` mantém valores ilustrativos e não contém segredos.

## Fluxo implementado

- `POST /v1/auth/register`: cria usuário, aplica hash seguro à senha, inicia sessão e envia a confirmação de e-mail.
- `POST /v1/auth/login`: autentica e cria uma nova sessão.
- `POST /v1/auth/logout`: revoga a sessão atual.
- `POST /v1/auth/resend-verification` e `POST /v1/auth/verify-email`: reenviam e consomem um link de confirmação de uso único.
- `POST /v1/auth/forgot-password` e `POST /v1/auth/reset-password`: iniciam e concluem a redefinição segura da senha.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico do ciclista.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.
- `GET /v1/onboarding`: consulta limitações, objetivos, disponibilidade e contexto opcional de ciclismo.
- `PUT /v1/onboarding/limitations`: salva informações de segurança.
- `PUT /v1/onboarding/goals`: salva até dois objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva a disponibilidade semanal.
- `PUT /v1/onboarding/cycling-context`: salva histórico resumido (horas, pedais, distância semanal recente, semanas de regularidade e maior distância), preferências de sessão, equipamento, terreno e meta opcional de prova.
- `GET /v1/assessments/current` e `POST /v1/assessments/submaximal`: consultam e registram o pedal de referência submáximo.
- `GET /v1/recovery/today` e `PUT /v1/recovery/today`: consultam e salvam o check-in diário de sono, estresse e fadiga percebida.
- `GET /v1/evolution/summary`: retorna totais observados, oito semanas de duração e métricas de pedal registradas, além de check-ins recentes para o atleta autenticado.
- `POST /v1/plans/generate`: gera e substitui o rascunho atual de quatro semanas.
- `GET /v1/plans/current`: consulta o plano ativo ou rascunho mais recente.
- `POST /v1/plans/{planID}/activate`: aprova um rascunho e mantém somente um plano ativo por atleta.
- `POST /v1/workouts/{workoutID}/start`: inicia uma sessão planejada do plano ativo.
- `POST /v1/workouts/{workoutID}/complete`: conclui a sessão e registra RPE, dificuldade, fadiga, dor e, opcionalmente, distância, elevação, frequência cardíaca e potência.
- `POST /v1/workouts/{workoutID}/explanation`: solicita uma explicação em linguagem simples; quando a IA está desligada ou indisponível, retorna o resumo validado pelo motor.
- `POST /v1/workouts/{workoutID}/cancel`: cancela uma sessão em andamento e mantém esse histórico.
- `GET /v1/activities`: lista, para o atleta autenticado, as sessões concluídas e canceladas.
- `POST /v1/feedback`: registra, para o atleta autenticado, uma experiência, problema ou sugestão com nota e mensagem. Os relatos pendentes podem entrar no resumo semanal do proprietário.

As sessões são opacas, armazenadas no PostgreSQL apenas como hash e enviadas ao navegador em cookie `HttpOnly`. Em produção, `APP_ENV=production` ativa também a exigência de HTTPS no cookie. Os links de confirmação e redefinição são aleatórios, expiram e só têm o hash armazenado; a redefinição de senha revoga todas as sessões existentes. A geração e a ativação de planos exigem e-mail confirmado.

As rotas atuais do frontend são `/`, `/entrar`, `/perfil`, `/plano`, `/atividades`, `/avaliacao`, `/recuperacao`, `/evolucao` e `/feedback`. A tela de atividades apresenta sessões concluídas e canceladas com data, duração, RPE e feedback. A aba de feedback de produto permite que atletas autenticados registrem a experiência, um problema ou uma sugestão; o relato é salvo no PostgreSQL sem expor o e-mail na resposta. Um job separado pode consolidar os relatos ainda não enviados em um resumo semanal pelo Resend, destinado somente ao endereço administrativo configurado na VPS. O perfil possui quatro etapas e retoma dados já salvos. Configure `frontend/.env` a partir de `frontend/.env.example` quando a URL da API for diferente de `http://localhost:8080`.

A tela `/plano` gera, apresenta e ativa ciclos de quatro semanas. O motor `rules-v1` é determinístico: considera experiência, objetivo, limitações, disponibilidade, o contexto opcional de ciclismo e um resumo observado dos últimos 28 dias de sessões e recuperação. Ele seleciona sessões específicas de forma gradual (cadência no indoor, subidas, sweet spot por potência/FTP, ritmo de prova e o piloto local de intervalos moderados de estrada), limita cada sessão ao tempo informado e reduz a intensidade quando há uma condição de segurança ativa ou sinais recentes de recuperação insuficiente. O dashboard usa o plano aprovado, explica a escala RPE e permite acompanhar a sessão do início ao feedback pós-treino. No desenvolvimento local, novos rascunhos também congelam no `prescription_snapshot` uma classificação observacional de prontidão e medições de 7/28/42 dias de aderência e carga por session-RPE; esses campos ainda não alteram a prescrição.

O feedback de uma sessão concluída adapta de forma conservadora os próximos treinos planejados. Dor, fadiga, dificuldade e diferença entre RPE planejado e realizado podem reduzir duração ou esforço; uma resposta claramente fácil permite somente uma progressão pequena de duração. A decisão fica registrada no treino e é apresentada na interface. As regras completas estão em `docs/training-adaptation-rules.md`.

A camada de IA explicativa é opcional e fica desligada por padrão. Em produção, o backend usa temporariamente a rota protegida `/cadencia/explanation` do Worker Cloudflare, que foi validada com o modelo Groq `openai/gpt-oss-20b`, para preservar a capacidade da VPS. O Ollama local permanece instalado, mas parado após uma medição de capacidade; a chamada ocorre somente no backend, nunca diretamente pelo navegador. Se os provedores não responderem, o usuário continua recebendo a explicação determinística do motor.

A rota `/avaliacao` permite registrar opcionalmente um pedal de referência submáximo, sem teste máximo ou diagnóstico. Para atletas avançados com objetivo de desempenho/prova, sem limitação ativa e com tempo suficiente, uma referência apta libera apenas intervalos controlados nas semanas de construção; não libera sprints nem esforço máximo.

A rota `/recuperacao` registra o check-in diário. Um sinal desfavorável gera cautela; fadiga máxima ou a combinação de dois sinais desfavoráveis indica necessidade de recuperação. Nesses casos, somente a próxima sessão futura do plano ativo pode ter duração e RPE reduzidos. Um check-in favorável mantém o plano e nunca aumenta a carga por si só. A decisão fica registrada no treino para não aplicar a mesma redução duas vezes.

A rota `/evolucao` organiza o que foi registrado: sessões concluídas e canceladas, tempo e distância por semana, elevação acumulada, médias opcionais de potência e frequência cardíaca, RPE, consistência e check-ins recentes. Ela mostra somente dados observados e explicita quando ainda não há histórico suficiente; não estima desempenho físico nem faz diagnóstico.

Ao concluir um treino, o atleta pode acrescentar distância e ganho de elevação. Quem informa no perfil que usa sensor de frequência cardíaca ou medidor de potência recebe também os respectivos campos opcionais. Esses dados aparecem no resultado da sessão, no histórico e, quando preenchidos, de forma agregada na área de evolução.

Quando não existem mais sessões planejadas ou em andamento, o PostgreSQL marca o plano ativo como concluído. O usuário pode então gerar um novo ciclo sem apagar o histórico anterior. As regras de datas e estados estão em `docs/training-cycle-lifecycle.md`.

## Estado atual e próximas etapas

O MVP de ciclismo está publicado em produção real:

- Frontend: <https://cadencia.devsaulo.com.br>
- API: <https://cadencia-api.devsaulo.com.br>
- Produção implantada na VPS Oracle no commit `c768ef7` (`chore: registra novidades do catalogo de ciclismo`), com a correção da API em `5fbc668` e a versão do produto `0.7.0`.
- PostgreSQL permanece privado na rede Docker; o Cloudflare Tunnel expõe somente frontend e API.
- Cadastro, confirmação de e-mail, recuperação de senha, onboarding, plano, treino, feedback, adaptação, atividades, evolução e logout foram validados.
- Dependabot está com 0 alertas abertos; os testes Go, build Docker e `govulncheck` passaram.
- A aba `/feedback`, o endpoint `POST /v1/feedback` e o job de resumo semanal estão implementados e publicados; as migrações `000013`, `000014` e `000015` foram aplicadas na produção.
- O ajuste responsivo dos períodos nos gráficos da Evolução foi publicado e validado no domínio oficial; a rolagem horizontal interna agora preserva os rótulos no celular.
- A produção permanece em `c768ef7`, com a API funcional de `5fbc668`. O checkout local parte do commit `b0f32b7`, que registra o histórico observacional `training-history-v2`; a evolução local para `training-history-v3`, com períodos não sobrepostos, ainda está sem commit nem deploy. O catálogo inicial e o piloto de estrada estão publicados, sujeitos aos critérios de elegibilidade documentados.
- Toda atualização com funcionalidade visível deve atualizar `frontend/lib/release.ts` (`APP_VERSION` e `UPDATE_NOTES`) para que a novidade seja exibida na tela de primeiro acesso após a atualização. O modal é mostrado uma vez por conta, versão e navegador.

A restauração completa do backup em ambiente isolado já foi concluída. Ainda falta definir a cópia externa dos backups, monitoramento e hardening das portas dos outros aplicativos hospedados na VPS. O ajuste visual da mensagem de privacidade e da altura da tela inicial desktop também está registrado.

A publicação oficial do Cadência é feita somente pela composição Docker da VPS, com `cadencia.devsaulo.com.br` e `cadencia-api.devsaulo.com.br` no Cloudflare Tunnel dedicado. Uma cópia privada criada acidentalmente no Sites durante uma tentativa de deploy foi excluída; o Sites não faz parte do fluxo de produção.

Depois da publicação, a operação aguarda os primeiros relatos reais pela aba `/feedback` e a confirmação da entregabilidade do resumo semanal na segunda-feira. O fluxo de feedback, a adaptação de recuperação e a latência, os limites e o fallback da explicação pelo Worker já foram testados. A próxima fase do produto segue o roadmap de `melhorias.md`: prontidão, evolução versionada das regras, adaptação em ciclo fechado, integridade dos dados, segurança, feedback e auditabilidade; depois, ampliação criteriosa do catálogo de ciclismo. O escopo permanece exclusivo de ciclismo nesta fase. O Ollama já está instalado e testado, mas permanece desligado por consumo elevado na VPS. Consulte `docs/README.md` e `docs/project-status.md` para a ordem completa.

## Licença

Este projeto é distribuído sob a licença [MIT](LICENSE).

## PWA

O frontend inclui manifesto, ícones, suporte à instalação e uma tela offline segura. O service worker armazena somente recursos estáticos; respostas da API e dados autenticados nunca entram no cache offline.

A instalação e o comportamento autenticado do PWA foram validados em um celular por uma origem HTTPS temporária do Cloudflare Tunnel. Essa exposição foi criada somente para o teste, não incluiu o PostgreSQL e foi removida após a validação.

O registro do service worker ocorre apenas no build de produção. Para testar localmente, pare o servidor de desenvolvimento, execute `npm run build` e depois `npm run preview:pwa` dentro de `frontend/`. A prévia usa a porta 3000, já autorizada pela API local. A instalação exige HTTPS ou `localhost`/`127.0.0.1`.
