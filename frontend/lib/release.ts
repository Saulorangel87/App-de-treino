import type { Locale } from './i18n';
import { UPDATE_NOTES_EN } from './release.en';

export const APP_VERSION = '0.39.0';

export type UpdateNote = Readonly<{
  version: string;
  title: string;
  description: string;
}>;

export const UPDATE_NOTES: readonly UpdateNote[] = [
  {
    version: '0.39.0',
    title: 'O resumo da sua semana',
    description:
      'Na tela de Evolução, um cartão novo conta como foi a sua última semana completa: quantos treinos você concluiu, o tempo de pedal comparado com a semana anterior, a zona média, a fadiga e se houve dor. O texto é montado pelas regras do app a partir dos seus registros, sem enviar nada para fora, e não é diagnóstico.',
  },
  {
    version: '0.38.0',
    title: 'O Cadência agora também fala inglês',
    description:
      'No topo de cada tela, e também em Configurações, você escolhe entre português e inglês. A escolha vale para as telas, os treinos e as explicações, os avisos e os e-mails. Os Termos de Uso e a Política de Privacidade também ganharam uma tradução; a versão em português continua sendo a que vale juridicamente. O idioma fica guardado no aparelho em que você escolheu.',
  },
  {
    version: '0.37.0',
    title: 'Treinos que acompanham a sua recuperação',
    description:
      'Quando você relata dor ou fadiga alta, o plano agora se protege em níveis (leve, moderado ou forte) em vez de travar o ciclo inteiro. A proteção tem prazo, diminui conforme os dias passam e é reavaliada a cada treino concluído e a cada check-in. Na tela do plano, o aviso explica o motivo e até quando vale, e o botão "Estou recuperado" pede uma nova avaliação na hora. Se houve dor nos últimos dias, a proteção é mantida por segurança.',
  },
  {
    version: '0.36.0',
    title: 'Vincule a atividade importada ao treino',
    description:
      'Em Atividades > Importar, cada atividade agora pode ser vinculada ao treino planejado do mesmo dia (ou de um dia próximo), e o vínculo aparece na lista com o nome do treino. O vínculo só guarda a relação: não altera o plano nem o treino. Também dá para vincular atividades que você já tinha importado e desfazer o vínculo quando quiser.',
  },
  {
    version: '0.35.0',
    title: 'Um visual novo, com cara de mapa de trilha',
    description:
      'O Cadência ganhou identidade própria inspirada nas cartas topográficas de MTB e gravel. O treino do dia aparece desenhado como um percurso, e a intensidade usa a sinalização de trilhas: círculo verde para leve, quadrado azul para moderado e losango preto para intenso. A navegação agora é a mesma em todas as telas e, no celular, fica numa barra inferior ao alcance do polegar.',
  },
  {
    version: '0.34.0',
    title: 'Importe suas atividades reais',
    description:
      'Em Atividades, agora dá para importar o arquivo .fit ou .gpx do seu relógio ou ciclocomputador (Garmin, Wahoo, XOSS e outros). O Cadência mostra o resumo — duração, distância, elevação, frequência cardíaca, potência e cadência — e sugere o treino planejado no mesmo dia; você confirma se quer usar esses dados ao concluir a sessão. No Android, instale o Cadência na tela inicial para compartilhar o arquivo direto do app do seu aparelho.',
  },
  {
    version: '0.33.0',
    title: 'Mais segurança no acesso à conta',
    description:
      'Em Configurações você agora pode trocar a senha e encerrar o acesso em outros aparelhos. Ao alterar a senha, os demais dispositivos são desconectados automaticamente. Também tornamos o Cadência mais estável: pedidos de rede lentos são repetidos com segurança e, se sua sessão expirar, você é levado de volta ao login.',
  },
  {
    version: '0.32.0',
    title: 'Onboarding de ciclismo mais completo e seguro',
    description:
      'O perfil agora pode registrar contexto adicional de ciclismo, disponibilidade, histórico, equipamento, FTP e sinais de segurança. O questionário adaptativo apresenta cada pergunta no momento certo, enquanto medidas opcionais e indicadores de evolução permanecem observacionais e não alteram o plano sem evidência suficiente.',
  },
  {
    version: '0.31.0',
    title: 'Configurações e encerramento seguro da conta',
    description:
      'A nova área de configurações reúne o acesso ao perfil e os dados da conta. Se você decidir encerrar sua conta, o Cadência exige sua senha e uma confirmação explícita antes de apagar permanentemente seus dados do banco de dados.',
  },
  {
    version: '0.30.0',
    title: 'Proteção após prova e adaptação mais íntegra',
    description:
      'Depois de um evento informado, o plano pode priorizar até sete dias de recuperação leve, com duração e RPE conservadores. O registro do pedal também aceita cadência média como dado opcional de observação. Agora cada sessão também mostra os dados, restrições e alternativas considerados pela regra. Feedback incompleto, dados inválidos e sinais protetivos recentes deixam de alterar a carga automaticamente; limitações e retorno após pausa continuam prioritários.',
  },
  {
    version: '0.29.0',
    title: 'Contexto de segurança e equipamento',
    description:
      'O feedback pós-treino agora pode registrar o equipamento utilizado. O perfil também permite informar sintomas de alerta e restrição médica; esses dados reforçam os bloqueios de segurança e não substituem avaliação profissional.',
  },
  {
    version: '0.28.0',
    title: 'Piloto de limiar controlado',
    description:
      'Atletas avançados elegíveis podem indicar Limiar como preferência para uma sessão conservadora de três blocos. O Cadência exige contexto de estrada ou indoor, avaliação apta e histórico mínimo; a sessão usa RPE, não define potência universal e mantém todas as proteções de dor e recuperação.',
  },
  {
    version: '0.27.0',
    title: 'Situação de treino mais explícita',
    description:
      'No perfil, você pode informar se está treinando regularmente ou retornando após uma pausa. Apenas a opção de retorno ativa sessões leves e contínuas de até 45 minutos em RPE 3,5; semanas informadas, sozinhas, não reduzem o plano.',
  },
  {
    version: '0.26.0',
    title: 'Catálogo com pedal longo explícito',
    description:
      'A sessão mais longa do ciclo agora aparece como Pedal longo, com estrutura de endurance e carga conservadora. O nome diferencia o objetivo sem aumentar duração, RPE ou frequência automaticamente.',
  },
  {
    version: '0.25.0',
    title: 'Feedback pós-treino mais completo',
    description:
      'Ao concluir um pedal, você pode registrar satisfação, terreno e condições externas em campos estruturados. Esses dados ficam disponíveis no histórico e continuam observacionais, sem alterar automaticamente o plano.',
  },
  {
    version: '0.24.0',
    title: 'Correção segura de métricas do pedal',
    description:
      'Registros pós-treino marcados para revisão agora podem ter suas métricas opcionais corrigidas sem alterar duração, RPE, feedback ou o plano. O valor original fica preservado para auditoria.',
  },
  {
    version: '0.23.0',
    title: 'Registros inconsistentes mais claros',
    description:
      'Quando o tempo e as métricas do pedal não combinam, o registro continua preservado, mas o app avisa que ele não será usado na observação do histórico.',
  },
  {
    version: '0.22.0',
    title: 'Erros de acesso mais claros',
    description:
      'Problemas temporários da API ou do banco agora aparecem como erro de carregamento nas telas autenticadas. O Cadência só volta ao login quando a sessão realmente não está autenticada.',
  },
  {
    version: '0.21.0',
    title: 'Contexto de segurança mais completo',
    description:
      'Ao informar uma limitação, você pode registrar localização, intensidade percebida, o que agrava e quando começou. Esses dados ajudam a preservar uma leitura mais segura e não substituem avaliação profissional.',
  },
  {
    version: '0.20.0',
    title: 'Perfil sempre acessível no menu',
    description:
      'A navegação lateral agora se ajusta melhor a telas com pouco espaço e mantém o acesso ao seu perfil fora da área coberta pelo rodapé.',
  },
  {
    version: '0.19.0',
    title: 'Histórico de novidades mais organizado',
    description:
      'O aviso inicial ficou mais curto e agora você pode consultar todas as atualizações na nova página de novidades, também disponível no menu.',
  },
  {
    version: '0.18.0',
    title: 'Feedback pós-treino com mais contexto',
    description:
      'Ao concluir um treino, você pode registrar sua recuperação percebida e a confiança para repetir a sessão. Esses sinais são observacionais e ainda não alteram automaticamente a carga.',
  },
  {
    version: '0.17.0',
    title: 'Registro de conclusão parcial',
    description:
      'Ao concluir um treino, você pode indicar se fez toda a sessão ou apenas parte dela e informar o motivo. Esse contexto ajuda o app a interpretar o realizado sem aumentar a carga automaticamente.',
  },
  {
    version: '0.16.0',
    title: 'Escopo de modalidades mais claro',
    description:
      'O Cadência mantém o foco em estrada, MTB XCO, gravel, XCM e indoor. Downhill/enduro e pista sprint/BMX não fazem parte deste app e não são aceitos como modalidades de treino.',
  },
  {
    version: '0.15.0',
    title: 'Piloto de intervalos curtos autorregulados',
    description:
      'Atletas avançados elegíveis que escolherem intervalos curtos podem receber, na estrada ou no indoor, seis repetições controladas de 1 minuto com recuperação leve. O app mantém RPE, histórico mínimo e as proteções de dor e recuperação, sem liberar sprint máximo.',
  },
  {
    version: '0.14.0',
    title: 'Piloto de intervalos VO₂max para estrada',
    description:
      'Atletas avançados elegíveis que escolherem VO₂max podem receber quatro blocos controlados de 4 minutos, baseados em evidências recentes. O app mantém RPE, duração mínima e as proteções de dor, recuperação e histórico.',
  },
  {
    version: '0.13.0',
    title: 'Taper pré-prova orientado por evento',
    description:
      'Atletas avançados elegíveis com evento próximo podem receber uma redução conservadora do volume, mantendo a frequência planejada e as proteções de dor e recuperação.',
  },
  {
    version: '0.12.0',
    title: 'Planejamento por proximidade do evento',
    description:
      'Quando há um evento futuro informado, a fase específica só é usada no período próximo à prova; eventos distantes mantêm a progressão regular.',
  },
  {
    version: '0.12.0',
    title: 'Sessões adaptadas iniciáveis com mais segurança',
    description:
      'Uma sessão ajustada por recuperação pode ser iniciada normalmente, enquanto as proteções de dor, fadiga e recuperação continuam valendo.',
  },
  {
    version: '0.12.0',
    title: 'Proteções de acesso e feedback',
    description:
      'O app reforça a proteção das rotas de autenticação, evita processamento concorrente do resumo semanal e só marca as novidades depois que você as dispensa.',
  },
  {
    version: '0.11.1',
    title: 'Semana de recuperação sem sessão de qualidade',
    description:
      'A quarta semana do ciclo preserva o pedal longo e usa recuperação ativa nas demais sessões, sem ritmo de prova ou intervalos de qualidade.',
  },
  {
    version: '0.11.0',
    title: 'Piloto de intervalos intensos para estrada',
    description:
      'Atletas avançados elegíveis podem receber, em ciclos alternados, uma sessão de intervalos intensos baseada em evidências recentes. O app mantém limites conservadores e as proteções de dor e recuperação.',
  },
  {
    version: '0.10.0',
    title: 'Registro de treino não realizado',
    description:
      'Treinos passados que não foram feitos podem ser registrados explicitamente. O app não cria reposição automática nem aumenta a carga seguinte por causa desse registro.',
  },
  {
    version: '0.9.0',
    title: 'Recuperação ativa na semana de recuperação',
    description:
      'O plano pode alternar uma sessão de recuperação ativa, com esforço leve e volume reduzido, sem substituir as proteções aplicadas quando há dor, limitação ou recuperação insuficiente.',
  },
  {
    version: '0.8.0',
    title: 'Piloto de intervalos aeróbicos para MTB XCO',
    description:
      'Atletas avançados com avaliação submáxima apta, objetivo compatível e contexto XCO explícito podem receber um piloto aeróbico com blocos controlados, sem sprint máximo ou simulação técnica de prova.',
  },
  {
    version: '0.7.0',
    title: 'Catálogo de ciclismo baseado em evidências',
    description:
      'As referências científicas agora são organizadas por modalidade e sustentam protocolos elegíveis, começando pelo piloto de intervalos moderados de estrada.',
  },
  {
    version: '0.6.0',
    title: 'Contexto observado no plano',
    description:
      'Ao gerar um novo plano, você consegue ver quais registros recentes ajudaram a definir uma progressão mais adequada.',
  },
  {
    version: '0.6.0',
    title: 'Feedback e recuperação com mais contexto',
    description:
      'Seus treinos concluídos, esforço percebido e check-ins de recuperação ajudam a manter as próximas sessões mais seguras.',
  },
  {
    version: '0.6.0',
    title: 'Registro do pedal',
    description:
      'Você pode registrar distância, elevação, frequência cardíaca e potência junto do treino, quando tiver esses dados.',
  },
  {
    version: '0.6.0',
    title: 'Treinos explicáveis',
    description:
      'Cada sessão continua mostrando sua estrutura, o motivo da escolha e os cuidados importantes para executar o treino.',
  },
] as const;

/** Notas de versão no idioma pedido (o inglês fica em release.en.ts, na mesma ordem). */
export function updateNotes(locale: Locale): readonly UpdateNote[] {
  return locale === 'en' ? UPDATE_NOTES_EN : UPDATE_NOTES;
}
