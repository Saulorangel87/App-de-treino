# Idiomas: português e inglês

O Cadência funciona em português (padrão) e em inglês. A escolha fica no cookie `cadencia_lang` (`pt` ou `en`), por aparelho, e pode ser feita no login, em Configurações, no rodapé e nas páginas de Termos e Privacidade. Sem cookie, tudo fica em português; o idioma do navegador não muda nada.

## Como funciona

**Frontend.** O layout lê o cookie no servidor (`lib/server-locale.ts`) e já renderiza a página no idioma certo: `<html lang>`, título, descrição e textos, sem piscar em português. Cada componente guarda os próprios textos nas duas línguas com `defineMessages({ pt: {...}, en: {...} })` (`lib/i18n.ts`) e os lê com `useMessages`. O TypeScript exige que o inglês tenha as mesmas chaves do português. Trocar de idioma grava o cookie e recarrega a página, para telas e dados da API mudarem juntos. Datas e números seguem o idioma (`INTL_LOCALE`, `formatDecimal`).

**API.** O frontend envia `Accept-Language` com o idioma do app. O middleware `withLanguage` guarda esse idioma na requisição. O `writeJSON` traduz a resposta na saída: percorre o JSON e troca cada texto em português que tenha entrada no catálogo `backend/internal/i18n/catalog_en.go`. Códigos (`error.code`, `status`, chaves) nunca mudam. O motor, as regras e o banco continuam em português; nenhuma regra de prescrição depende do idioma.

O catálogo reconhece:

- frases exatas (nome, objetivo, etapas, resumo, regras, base científica, razões da proteção e da adaptação, erros e mensagens da API);
- frases com números ou rótulos (`englishPatterns`, como `"Bloco de limiar %d de %d"`), em que um `%s` só vale se também tiver tradução, para o texto do próprio atleta nunca casar com um formato genérico;
- frases que o motor monta juntando duas frases conhecidas (resumo + nota do taper, por exemplo).

Texto sem entrada fica como está: anotações do atleta e planos gerados por versões antigas do motor, com frases que não existem mais.

**E-mails.** Confirmação de e-mail e redefinição de senha saem no idioma da requisição que os pediu (`httpapi/language.go`). O resumo semanal de feedback é só para o administrador e continua em português.

**Planilha de dados.** Abas, cabeçalhos e rótulos fixos saem traduzidos; datas continuam no formato brasileiro, como o cabeçalho avisa.

**Explicação por IA.** O Ollama local e o Worker da Cloudflare respondem no idioma pedido. O código do Worker não está neste repositório (fica no painel da Cloudflare, Worker `flat-rice-6724`); desde 08/10/2026 ele aceita o campo `language`, que a API só envia em inglês, então o contrato em português não mudou. Se o Worker for recriado, ele precisa manter esse campo, a mensagem de sistema em inglês e `buildCadenciaPromptEn`; sem isso, em inglês ele recusa e a API devolve a explicação das próprias regras, traduzida.

**Documentos legais.** Termos e Política de Privacidade têm tradução em `frontend/lib/legal.en.ts`, com aviso de que a versão em português é a que vale juridicamente. Um teste confere que as seções têm os mesmos ids, na mesma ordem e com o mesmo número de parágrafos.

## Ao mudar um texto

- **Tela:** escreva as duas línguas no `defineMessages` do componente.
- **Mensagem de erro ou `message` da API:** acrescente a frase ao catálogo. `TestEveryAPIMessageHasAnEnglishEntry` lê o código do `httpapi` e falha se faltar.
- **Texto do motor que o atleta vê:** acrescente ao catálogo. `TestEveryWorkoutTextTheAthleteReadsHasAnEnglishEntry` gera planos para muitos perfis e falha se alguma frase ficar sem inglês.
- **Nota de versão:** escreva também em `frontend/lib/release.en.ts`, na mesma posição.
- **Termos ou privacidade:** atualize `legal.en.ts` no mesmo PR.
