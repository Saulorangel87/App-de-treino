export const APP_VERSION = '0.19.0';

export type UpdateNote = Readonly<{
  version: string;
  title: string;
  description: string;
}>;

export const UPDATE_NOTES: readonly UpdateNote[] = [
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
