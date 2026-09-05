# Infraestrutura

Esta pasta contém a configuração versionada de produção do Cadência na VPS Oracle. A composição Docker, o Cloudflare Tunnel dedicado, as migrações controladas e as rotinas de backup já estão implantados e validados.

Nenhuma credencial de produção deve ser armazenada no repositório. O arquivo `.env.production` existe somente na VPS e contém os segredos do PostgreSQL, Resend e Cloudflare.

O último deploy oficial documentado é o commit `c768ef7`, publicado na VPS Oracle depois da atualização funcional `5fbc668`. A API e o frontend foram reconstruídos, a migração `000015` do catálogo foi aplicada e, em seguida, o frontend foi recriado novamente para publicar a versão `0.7.0` e sua nota de novidades. Consulte `cadencia/README.md` para o primeiro deploy, atualizações, backup e restauração. A situação operacional atual, o inventário de serviços e as pendências de firewall estão em `docs/project-status.md`. O ambiente Sites não faz parte do fluxo de produção do Cadência.
