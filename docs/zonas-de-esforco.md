# Zonas de esforço

O app descreve a intensidade do treino em **cinco zonas (Z1 a Z5)**, e não mais em RPE. O motor, porém, continua prescrevendo, adaptando e guardando em RPE (escala de 1 a 10). As zonas são a linguagem do atleta; o RPE é a escala interna, e uma tabela liga as duas.

## Por que as zonas são uma camada, e não um substituto

O RPE está em todo o sistema: no que o planejador prescreve, na adaptação de curto prazo (o gatilho do banco compara o esforço informado com o esperado), na proteção graduada, nas evidências científicas cadastradas (que se apoiam no esforço percebido da sessão), na carga de treino da aba Evolução e nos dados já guardados. Trocar a variável do motor seria uma reescrita grande, sem ganho para o atleta. Então o atleta vê zonas, e o servidor converte.

## A tabela

| Zona | Nome | RPE prescrito | RPE gravado ao informar a zona | Teste da conversa |
| --- | --- | --- | --- | --- |
| Z1 | Recuperação | até 3,5 | 3 | Conversa fácil, respiração tranquila, pernas leves |
| Z2 | Resistência | de 3,6 a 5 | 4,5 | Frases completas; sustentável por horas |
| Z3 | Ritmo | de 5,1 a 6,5 | 6 | Só frases curtas; firme, mas controlado |
| Z4 | Limiar | de 6,6 a 7,5 | 7 | Poucas palavras; sustentável por 20 a 40 minutos |
| Z5 | Intenso | acima de 7,5 | 8,5 | Não dá para conversar; poucos minutos |

Todo RPE que o planejador usa cai na zona prevista: recuperação e giro leve protegido (3,5) em Z1; giro de base (4), endurance e pedal longo (5) em Z2; tempo e intervalos moderados (6) e subidas controladas (6,5) em Z3; sweet spot (7) e intervalos de limiar (7,5) em Z4; intervalos intensos (8) em Z5. Um teste confere isso, e outro confere que a tabela do servidor (`backend/internal/zones/zones.go`) é igual à da tela (`frontend/lib/zones.ts`).

Os limites entre as zonas são **escolha de produto**, como os limiares da proteção graduada. Não vêm de um estudo, e o app diz isso.

## Faixas em batimentos e watts

Quando o atleta informa a **frequência cardíaca máxima** (campo opcional no perfil, só com "Uso frequência cardíaca" marcado, entre 100 e 230 bpm) e/ou o **FTP**, cada zona aparece também como faixa: batimentos em percentual da frequência máxima (50–60, 60–70, 70–80, 80–90 e 90–100%) e watts em percentual do FTP (até 55, 55–75, 75–90, 90–105 e acima de 105%). As zonas vizinhas não se sobrepõem. Sem esses números, o app mostra só o teste da conversa e sugere informar a frequência máxima ou o FTP no perfil.

São referências, não prescrições. A fórmula "220 menos a idade" erra bastante, a frequência cardíaca atrasa em intervalos curtos e sobe com calor, cansaço e cafeína. Para intervalos curtos, a potência é mais confiável.

## Como o feedback funciona

O formulário de conclusão pergunta "Em que zona você pedalou?" com cinco cartões (a zona planejada vem marcada). O app envia `actual_zone`, e o servidor grava o RPE que representa a zona. Quem ainda envia `actual_rpe` continua funcionando (clientes antigos); com os dois, vale a zona. A dificuldade e a fadiga depois do treino seguem sendo perguntadas à parte e carregam o "como foi", independente da zona.

O RPE gravado de cada zona fica dentro dela e perto do que o planejador prescreve, então uma sessão feita como planejada não dispara ajuste por engano. A adaptação reage quando o esforço informado fica 2 pontos de RPE acima ou abaixo do prescrito, ou com dificuldade e fadiga altas; como cada zona vale cerca de 1,5 ponto, um salto de duas zonas sempre chega lá e, em alguns treinos, o de uma zona também (por exemplo, de um giro de base em RPE 4 para Z3).

## O que mudou na tela

Plano, Hoje, estrutura do treino, cartão de adaptação (só mostra a mudança de zona quando ela existe), Atividades, Evolução, Recuperação e a planilha de dados passaram a mostrar a zona. Os símbolos de trilha seguem as zonas: verde para Z1 e Z2, azul para Z3 e preto para Z4 e Z5.

## O que continua em RPE

- O motor, o banco, a API (`target_rpe`, `actual_rpe`) e as explicações científicas.
- A cópia completa dos dados, que guarda o valor bruto.
- O banco da Avaliação (`actual_rpe`): a tela pergunta a zona, e a API guarda o RPE correspondente, como nos treinos.

## Meus números e sugestões (Fase B, `0.40.0`)

- **Onde:** Configurações → "Meus números". Frequência máxima, limiar de frequência (LTHR) e FTP, com a data do FTP. O valor é gravado no contexto de ciclismo do perfil (mesmo dado do perfil, sem duplicação). O LTHR só existe com "Uso frequência cardíaca" marcado, fica entre 100 e 230 bpm e nunca passa da frequência máxima.
- **Zonas pelo limiar:** quando há LTHR, as faixas em batimentos usam % do limiar (Z1 até 81%, Z2 82–89%, Z3 90–93%, Z4 94–99%, Z5 acima de 99%, modelo de Friel/Coggan); sem LTHR, continuam em % da frequência máxima. As telas leem o contexto atual do atleta (`use-live-cycling-context.ts`), então a mudança vale na hora, sem gerar plano novo.
- **Sugestões:** `GET /v1/activities/reference-suggestions`. Frequência máxima = maior batimento visto nas atividades importadas (100 a 230 bpm); FTP = 95% da melhor média de 20 minutos seguidos, que o importador `.fit` passou a calcular (`best_20min_power_watts`, migração `000038`; atividades já importadas precisam ser reenviadas). Só sugere com pelo menos 3 atividades com o dado, e nada é gravado sozinho: o valor vai para o campo e o atleta confirma. Um esforço de 20 minutos em treino leve pode subestimar o FTP.
- **Não muda:** o motor segue em RPE; só o texto das zonas e as sugestões usam esses números.

## Próximas fases

- **Fase B (passos 1 e 3 na `0.40.0`):** os três números ficam em um só lugar, o cartão "Meus números" em Configurações, e as sugestões vêm das atividades importadas. Detalhes na seção abaixo. Os testes guiados (passo 2) seguem para depois, com avisos de segurança e decisão de produto.
- **Fase C:** ler as batidas do arquivo `.fit` importado e calcular o tempo em cada zona, para comparar o planejado com o pedalado. Só com arquivos do próprio atleta, sem dados do Strava.
- **Mais adiante:** o motor falar em zonas nativamente, quando houver dados reais que sustentem.
