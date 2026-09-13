# Infraestrutura

Esta pasta contém a configuração versionada de produção do Cadência na VPS Oracle. A composição Docker, o Cloudflare Tunnel dedicado, as migrações controladas e as rotinas de backup já estão implantados e validados.

Nenhuma credencial de produção deve ser armazenada no repositório. O arquivo `.env.production` existe somente na VPS e contém os segredos do PostgreSQL, Resend e Cloudflare.

O último deploy oficial documentado é o commit `84b653b`, publicado na VPS Oracle para publicar a versão `0.20.0`. A API e o frontend foram reconstruídos, as migrações `000020` e `000021` foram aplicadas e os serviços foram validados após o backup preventivo. A release correspondente ainda será registrada no GitHub. Consulte `cadencia/README.md` para o primeiro deploy, atualizações, backup e restauração. A situação operacional atual, o inventário de serviços e as pendências de firewall estão em `docs/project-status.md`. O ambiente Sites não faz parte do fluxo de produção do Cadência.
