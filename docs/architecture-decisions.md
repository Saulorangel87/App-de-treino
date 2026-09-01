# Decisões de arquitetura

Última revisão: 31 de agosto de 2026.

## ADR-001 — Banco de dados próprio

**Status:** Aceita

O Cadência usará PostgreSQL controlado pelo proprietário do projeto.

- Desenvolvimento e testes: PostgreSQL local, preferencialmente executado com Docker Compose.
- Produção: PostgreSQL hospedado na VPS Oracle do proprietário.
- O ambiente de desenvolvimento e o de produção usarão o mesmo mecanismo de banco e as mesmas migrações.
- A aplicação não dependerá do banco gerenciado da hospedagem do frontend.
- Endereços, usuários, senhas e certificados serão fornecidos por variáveis de ambiente e nunca serão gravados no repositório.
- A conexão de produção deverá usar TLS, usuário exclusivo da aplicação, privilégios mínimos, firewall restritivo e backups automatizados.
- O backend em Go será o único componente autorizado a acessar o banco. O navegador nunca receberá credenciais do PostgreSQL.

## Estratégia de ambientes

### Local

O repositório terá um `compose.yaml` com PostgreSQL para permitir subir um ambiente reproduzível. Os dados locais ficarão em volume separado e não serão enviados ao repositório.

### Produção

O backend em Go será executado na VPS ou em outro servidor autorizado a alcançar o PostgreSQL da VPS. A configuração usará uma variável `DATABASE_URL` específica do ambiente.

O projeto também utiliza Cloudflare Tunnel para expor serviços HTTP/HTTPS quando necessário. O túnel deve encaminhar tráfego somente para o frontend e/ou para a API, conforme a configuração do ambiente; ele não substitui nem expõe diretamente o PostgreSQL. As credenciais e o arquivo de configuração do túnel permanecem fora do repositório.

### Migrações

Todas as alterações de estrutura serão versionadas em arquivos SQL. O mesmo conjunto de migrações será aplicado primeiro localmente e depois na VPS, com backup antes de mudanças de produção.

## Fluxo de acesso

```text
Frontend React
      |
      | HTTPS / REST
      v
Backend Go
      |
      | conexão PostgreSQL protegida
      v
PostgreSQL local (desenvolvimento)
ou PostgreSQL na VPS Oracle (produção)
```

## Estado da implementação

Concluído localmente:

1. PostgreSQL 17 em Docker Compose, exposto somente em loopback.
2. Migrações SQL versionadas até `000011`, incluindo fontes científicas, contexto opcional de ciclismo, avaliações submáximas e ganho de elevação por sessão.
3. API REST em Go com verificações de saúde e prontidão.
4. Autenticação com bcrypt, sessões opacas, hash do token e cookie `HttpOnly`.
5. Frontend conectado exclusivamente à API.
6. Onboarding persistente, motor `rules-v1`, planos de quatro semanas e aprovação transacional.
7. Ciclo de vida das sessões, feedback pós-treino e adaptação conservadora das próximas cargas.
8. Encerramento automático do plano e geração do próximo ciclo sem apagar o histórico.
9. PWA, interface responsiva e validações locais de integração.
10. Contexto opcional e condicional de ciclismo persistido (histórico resumido, terreno, equipamento, FTP e meta de prova).
11. Avaliação inicial submáxima persistida, sem teste máximo ou diagnóstico.
12. Check-in diário persistido na tabela `recovery_data`, com data local enviada pelo cliente e adaptação conservadora, transacional e idempotente da próxima sessão.
13. Área de evolução baseada em dados observados: sessões encerradas, duração e distância semanais, elevação acumulada, médias opcionais de potência/frequência cardíaca, RPE, consistência e check-ins, sem pontuação de saúde, estimativa de performance ou diagnóstico.
14. Métricas do pedal são opcionais: distância e elevação para qualquer sessão; frequência cardíaca e potência somente quando o equipamento está informado no perfil. A elevação é persistida pela migração `000011`.
15. PWA instalado e validado em um celular real por origem HTTPS temporária, incluindo navegação autenticada, comportamento responsivo, gráficos, modal de sessão e logout. O túnel e seus registros DNS temporários foram removidos depois do teste.
16. Configuração de produção preparada localmente em `infrastructure/cadencia/`, com imagens Docker para frontend e API, PostgreSQL em rede interna, usuário de aplicação sem privilégios administrativos, migrações controladas, backup verificável e Cloudflare Tunnel dedicado. A configuração não foi implantada na VPS nem recebeu segredos reais.

## Próximas implementações

1. Implantar e validar a configuração preparada na VPS Oracle: banco privado, migrações, backup automatizado e teste de restauração. O Cloudflare Tunnel dedicado deve expor somente o frontend e a API.

Uma implantação só será considerada concluída depois de validar HTTPS, CORS, cookies seguros, migrações, backup e restauração.

