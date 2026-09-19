# Cadência

Aplicação de planejamento adaptativo de treinos de ciclismo.

Versão publicada: `0.32.0`. A entrega inclui questionário adaptativo versionado, contexto seguro ampliado do perfil/ciclismo e novos indicadores observacionais de evolução. A migração `000030_profile_safety_context` foi aplicada em produção após backup verificável.

O escopo do Cadência é ciclismo de estrada, MTB XCO, XCM, gravel e indoor. Sprint/pista/BMX e downhill/enduro não fazem parte deste app e não são aceitos como modalidades de treino.

## Estrutura

- `frontend/`: aplicação React/TypeScript publicada na VPS Oracle por Docker e Cloudflare Tunnel.
- `backend/`: API REST em Go.
- `database/`: migrações PostgreSQL versionadas.
- `api/`: contrato OpenAPI.
- `infrastructure/`: configuração versionada para a VPS Oracle.
- `docs/`: decisões, regras, ciclo de vida e catálogo científico do produto.
- `docs/README.md`: índice da documentação e regra de atualização.
- `planejamento.md`: documento único de visão, escopo, estado e próximos passos do projeto.

## Ambiente local

1. Copie `.env.example` para `.env` e use somente credenciais locais.
2. Inicie o PostgreSQL com `docker compose up -d postgres`.
3. Aplique os arquivos `database/migrations/*.up.sql` ainda pendentes, em ordem numérica. O esquema versionado inclui as migrações `000001`–`000030`; a produção está sincronizada até `000030`.
4. Execute a API com `pwsh -NoProfile -File scripts/run-api.ps1`.
5. Execute o frontend a partir de `frontend/` com `npm run dev`.

O script da API compila o binário em `backend/.gotmp`, pasta ignorada pelo Git, porque o Smart App Control do Windows pode bloquear o executável temporário criado pelo `go run`.

O frontend nunca se conecta diretamente ao PostgreSQL. Todo acesso passa pela API Go.
A configuração local deste projeto usa a porta `5433` no `.env`, pois a `5432` já estava ocupada no Windows. O `.env.example` mantém valores ilustrativos e não contém segredos.

## Fluxo implementado

- `POST /v1/auth/register`: cria usuário, aplica hash seguro à senha, inicia sessão e envia a confirmação de e-mail.
- `POST /v1/auth/login`: autentica e cria uma nova sessão.
- `POST /v1/auth/logout`: revoga a sessão atual.
- `DELETE /v1/auth/account`: exige a senha atual e a confirmação `ENCERRAR CONTA` para apagar a conta e todos os dados pessoais em cascata no PostgreSQL.
- `POST /v1/auth/resend-verification` e `POST /v1/auth/verify-email`: reenviam e consomem um link de confirmação de uso único.
- `POST /v1/auth/forgot-password` e `POST /v1/auth/reset-password`: iniciam e concluem a redefinição segura da senha.
- `GET /v1/me`: retorna o usuário autenticado.
- `GET /v1/profile`: consulta o perfil básico do ciclista.
- `PUT /v1/profile`: cria ou atualiza o perfil básico.
- `GET /v1/onboarding`: consulta limitações, objetivos, disponibilidade e contexto opcional de ciclismo.
- `GET /v1/onboarding/questionnaire`: expõe o contrato versionado de etapas, perguntas obrigatórias e gates condicionais do onboarding de ciclismo.
- `PUT /v1/onboarding/limitations`: salva informações de segurança, incluindo opcionalmente localização, intensidade percebida, movimento agravante, data de início, sintomas de alerta, restrição médica, cirurgia recente, proibição de exercício e condição que afeta exercício. Esses campos são contexto informado pelo atleta e não constituem diagnóstico.
- `PUT /v1/onboarding/goals`: salva até dois objetivos priorizados.
- `PUT /v1/onboarding/availability`: salva a disponibilidade semanal.
- `PUT /v1/onboarding/cycling-context`: salva histórico resumido (tempo de prática, horas, pedais, duração e distância recentes), preferências, GPS/relógio/rolo, sensores, FTP opcional com data/protocolo, potência média, terreno e meta opcional de prova com distância e data futura válidas.
- `GET /v1/assessments/current` e `POST /v1/assessments/submaximal`: consultam e registram o pedal de referência submáximo.
- `GET /v1/recovery/today` e `PUT /v1/recovery/today`: consultam e salvam o check-in diário de sono, estresse e fadiga percebida.
- `GET /v1/evolution/summary`: retorna totais observados, oito semanas de duração, carga sessão-RPE, velocidade média calculada de registros e acompanhamento factual dos objetivos nos últimos 28 dias, além de check-ins recentes.
- `POST /v1/plans/generate`: gera e substitui o rascunho atual de quatro semanas.
- `GET /v1/plans/current`: consulta o plano ativo ou rascunho mais recente.
- `POST /v1/plans/{planID}/activate`: aprova um rascunho e mantém somente um plano ativo por atleta.
- `POST /v1/workouts/{workoutID}/start`: inicia uma sessão planejada ou adaptada do plano ativo.
- `POST /v1/workouts/{workoutID}/complete`: conclui a sessão e registra RPE, dificuldade, fadiga, dor e, opcionalmente, distância, elevação, frequência cardíaca, potência, cadência média e equipamento utilizado; também registra `completion_status`, o `partial_reason` controlado quando necessário, recuperação percebida e confiança para repetir.
- `POST /v1/workouts/{workoutID}/correct`: corrige somente métricas opcionais de pedal de uma sessão concluída marcada como inconsistente ou incompleta; duração, RPE, feedback, plano e prescrição não são alterados, e os valores originais ficam no histórico de auditoria.
- `POST /v1/workouts/{workoutID}/explanation`: solicita uma explicação em linguagem simples; quando a IA está desligada ou indisponível, retorna o resumo validado pelo motor.
- `POST /v1/workouts/{workoutID}/cancel`: cancela uma sessão em andamento e mantém esse histórico.
- `GET /v1/activities`: lista, para o atleta autenticado, as sessões concluídas e canceladas.
- `POST /v1/feedback`: registra, para o atleta autenticado, uma experiência, problema ou sugestão com nota e mensagem. Os relatos pendentes podem entrar no resumo semanal do proprietário.

O início de uma sessão também revalida a segurança contra uma limitação cadastrada depois da geração do plano: sessões acima de RPE 4 são bloqueadas até que exista um novo plano protegido ou orientação profissional. Sessões protegidas em RPE 4 ou abaixo continuam iniciáveis.

As sessões são opacas, armazenadas no PostgreSQL apenas como hash e enviadas ao navegador em cookie `HttpOnly`. Em produção, `APP_ENV=production` ativa também a exigência de HTTPS no cookie. Os links de confirmação e redefinição são aleatórios, expiram e só têm o hash armazenado; a redefinição de senha revoga todas as sessões existentes. A geração e a ativação de planos exigem e-mail confirmado.

As rotas atuais do frontend são `/`, `/entrar`, `/perfil`, `/configuracoes`, `/plano`, `/atividades`, `/avaliacao`, `/recuperacao`, `/evolucao`, `/feedback` e `/novidades`. A tela de atividades apresenta sessões concluídas e canceladas com data, duração, RPE e feedback. A aba de feedback de produto permite que atletas autenticados registrem a experiência, um problema ou uma sugestão; o relato é salvo no PostgreSQL sem expor o e-mail na resposta. Um job separado pode consolidar os relatos ainda não enviados em um resumo semanal pelo Resend, destinado somente ao endereço administrativo configurado na VPS. O perfil possui quatro etapas e retoma dados já salvos. A área de configurações mostra os dados da conta, encaminha para o perfil e permite o encerramento definitivo com confirmação dupla. Configure `frontend/.env` a partir de `frontend/.env.example` quando a URL da API for diferente de `http://localhost:8080`.

A tela `/plano` gera, apresenta e ativa ciclos de quatro semanas. O motor `rules-v1` é determinístico: considera experiência, objetivo, limitações, disponibilidade, o contexto opcional de ciclismo e um resumo observado dos últimos 28 dias de sessões e recuperação. Ele seleciona sessões específicas de forma gradual (cadência no indoor, subidas, sweet spot por potência/FTP, ritmo de prova e os pilotos de intervalos moderados de estrada, intensos de estrada, VO₂max de estrada, intervalos curtos autorregulados e aeróbicos XCO), limita cada sessão ao tempo informado e reduz a intensidade quando há uma condição de segurança ativa ou sinais recentes de recuperação insuficiente. Os pilotos VO₂max e de intervalos curtos exigem preferência explícita, elegibilidade restrita, avaliação apta e histórico mínimo; foram publicados na versão `0.20.0` com as migrações correspondentes. O dashboard usa o plano aprovado, explica a escala RPE e permite acompanhar a sessão do início ao feedback pós-treino. No desenvolvimento local e em produção, novos rascunhos também congelam no `prescription_snapshot` uma classificação observacional de prontidão e medições de 7/28/42 dias de aderência e carga por session-RPE; esses campos ainda não alteram a prescrição.

O feedback de uma sessão concluída adapta de forma conservadora os próximos treinos planejados. Dor, fadiga, dificuldade e diferença entre RPE planejado e realizado podem reduzir duração ou esforço; uma resposta claramente fácil permite somente uma progressão pequena de duração. A decisão fica registrada no treino e é apresentada na interface. No desenvolvimento local e em produção, a explicação também registra em `adaptation_shadow` a cobertura observacional de recuperação percebida, confiança para repetir, a proveniência da avaliação e o gate de distribuição dos estímulos, sem alterar a prescrição. Novos rascunhos também registram a distribuição observacional de sessões de qualidade e sua proximidade em `training-history-v4`, além da auditoria observacional da estrutura do ciclo em `periodization-shadow-v1`; esses blocos não prescrevem carga. As regras completas estão em `docs/training-adaptation-rules.md`.

Quando uma sessão concluída é marcada para revisão por dados incompatíveis, o atleta pode corrigir as métricas opcionais do pedal pela tela do plano. A correção reavalia `data-integrity-v1`, preserva duração, RPE e feedback e não recalcula o plano. Cada alteração guarda os valores anteriores em `data_integrity_corrections`; a sessão só volta à observação histórica se os dados corrigidos forem coerentes. Essa operação continua sem autoridade sobre o `rules-v1`.

A camada de IA explicativa é opcional e fica desligada por padrão. Em produção, o backend usa temporariamente a rota protegida `/cadencia/explanation` do Worker Cloudflare, que foi validada com o modelo Groq `openai/gpt-oss-20b`, para preservar a capacidade da VPS. O Ollama local permanece instalado, mas parado após uma medição de capacidade; a chamada ocorre somente no backend, nunca diretamente pelo navegador. Se os provedores não responderem, o usuário continua recebendo a explicação determinística do motor.

A rota `/avaliacao` permite registrar opcionalmente um pedal de referência submáximo, sem teste máximo ou diagnóstico. Para atletas avançados com objetivo de desempenho/prova, sem limitação ativa e com tempo suficiente, uma referência apta libera apenas intervalos controlados nas semanas de construção; não libera sprints nem esforço máximo.

A rota `/recuperacao` registra o check-in diário. Um sinal desfavorável gera cautela; fadiga máxima ou a combinação de dois sinais desfavoráveis indica necessidade de recuperação. Nesses casos, somente a próxima sessão futura do plano ativo pode ter duração e RPE reduzidos. Um check-in favorável mantém o plano e nunca aumenta a carga por si só. A decisão fica registrada no treino para não aplicar a mesma redução duas vezes.

A rota `/evolucao` organiza o que foi registrado: sessões concluídas e canceladas, tempo e distância por semana, elevação acumulada, médias opcionais de potência e frequência cardíaca, carga sessão-RPE, velocidade média calculada, acompanhamento factual dos objetivos e check-ins recentes. Ela mostra somente dados observados e explicita quando ainda não há histórico suficiente; não estima desempenho físico nem faz diagnóstico.

Ao concluir um treino, o atleta pode acrescentar distância, ganho de elevação e cadência média. Quem informa no perfil que usa sensor de frequência cardíaca ou medidor de potência recebe também os respectivos campos opcionais. A cadência é armazenada somente como observação, aparece no resultado da sessão e no histórico e não cria uma meta individual nem altera a prescrição.

Quando não existem mais sessões planejadas ou em andamento, o PostgreSQL marca o plano ativo como concluído. O usuário pode então gerar um novo ciclo sem apagar o histórico anterior. As regras de datas e estados estão em `docs/training-cycle-lifecycle.md`.

## Estado atual e próximas etapas

O MVP de ciclismo está publicado e validado em produção:

- Frontend: <https://cadencia.devsaulo.com.br>
- API: <https://cadencia-api.devsaulo.com.br>
- Código publicado na linha de versão `0.32.0`, incluindo questionário adaptativo, contexto seguro ampliado e indicadores observacionais de evolução.
- Versão visível: `0.32.0`; migrações de banco aplicadas até `000030`.
- PostgreSQL permanece privado na rede Docker; o Cloudflare Tunnel expõe somente frontend e API.
- Cadastro, onboarding, plano, treino, feedback, adaptação, atividades, evolução, novidades e logout foram validados em produção. A nova área de configurações está implementada localmente e aguarda validação manual antes de qualquer deploy.
- `rules-v1` continua sendo a única fonte prescritiva. Os shadows permanecem observacionais.
- Dependabot está com 0 alertas abertos; `go test`, `go vet`, build e auditoria de dependências de produção passaram. `govulncheck` não está instalado.

O planejamento vigente está em [`planejamento.md`](planejamento.md). As atividades restantes são manutenção operacional, feedback real, cópia externa de backups, monitoramento, hardening da VPS, limpeza gradual do lint histórico e calibração científica antes de ampliar a autoridade do motor. Corrida e musculação estão fora deste projeto.

A publicação oficial usa somente a composição Docker da VPS com os domínios oficiais e o Cloudflare Tunnel dedicado. O ambiente Sites não faz parte da produção do Cadência.

## Licença

Este projeto é distribuído sob a licença [MIT](LICENSE).

## PWA

O frontend inclui manifesto, ícones, suporte à instalação e uma tela offline segura. O service worker armazena somente recursos estáticos; respostas da API e dados autenticados nunca entram no cache offline.

A instalação e o comportamento autenticado do PWA foram validados em um celular por uma origem HTTPS temporária do Cloudflare Tunnel. Essa exposição foi criada somente para o teste, não incluiu o PostgreSQL e foi removida após a validação.

O registro do service worker ocorre apenas no build de produção. Para testar localmente, pare o servidor de desenvolvimento, execute `npm run build` e depois `npm run preview:pwa` dentro de `frontend/`. A prévia usa a porta 3000, já autorizada pela API local. A instalação exige HTTPS ou `localhost`/`127.0.0.1`.
