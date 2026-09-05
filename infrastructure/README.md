# Infraestrutura

Esta pasta contém a configuração versionada de produção do Cadência na VPS Oracle. A composição Docker, o Cloudflare Tunnel dedicado, as migrações controladas e as rotinas de backup já estão implantados e validados.

Nenhuma credencial de produção deve ser armazenada no repositório. O arquivo `.env.production` existe somente na VPS e contém os segredos do PostgreSQL, Resend e Cloudflare.

O último deploy oficial documentado é o commit `9d8c624`, publicado na VPS Oracle para publicar a versão `0.8.0`. A API e o frontend foram reconstruídos, a migração `000016` do piloto XCO foi aplicada e os serviços foram validados após o backup preventivo. Consulte `cadencia/README.md` para o primeiro deploy, atualizações, backup e restauração. A situação operacional atual, o inventário de serviços e as pendências de firewall estão em `docs/project-status.md`. O ambiente Sites não faz parte do fluxo de produção do Cadência.
