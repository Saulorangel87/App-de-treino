// Textos jurídicos do Cadência. Descrevem o que o app faz hoje; cada afirmação
// aqui foi conferida no código ou na infraestrutura. Quem alterar o tratamento
// de dados (nova integração, novo provedor, novo dado coletado) deve atualizar
// estes textos e LEGAL_VERSION no mesmo PR.

export const LEGAL_VERSION = '2026-09-30';
export const LEGAL_UPDATED_LABEL = '30 de setembro de 2026';
export const LEGAL_CONTACT_EMAIL = 'sauloleonardo1987@gmail.com';

export type LegalSection = {
  id: string;
  title: string;
  paragraphs?: string[];
  items?: string[];
  after?: string[];
};

export type LegalDocument = {
  title: string;
  kicker: string;
  intro: string;
  sections: LegalSection[];
};

export const PRIVACY_POLICY: LegalDocument = {
  kicker: 'PRIVACIDADE',
  title: 'Política de Privacidade',
  intro:
    'Esta política explica quais dados o Cadência guarda sobre você, para que usa cada um, com quem compartilha e como você os controla. Ela segue a Lei Geral de Proteção de Dados (Lei 13.709/2018).',
  sections: [
    {
      id: 'responsavel',
      title: 'Quem é o responsável pelos seus dados',
      paragraphs: [
        'O Cadência é desenvolvido e operado por Saulo Rangel, que decide como os dados são tratados (controlador, na LGPD).',
        `Para qualquer pedido ou dúvida sobre seus dados, escreva para ${LEGAL_CONTACT_EMAIL}.`,
      ],
    },
    {
      id: 'dados',
      title: 'Quais dados coletamos',
      paragraphs: ['Coletamos apenas o que você informa ou gera ao usar o app:'],
      items: [
        'Conta: nome, e-mail, senha (guardada somente como hash irreversível, nunca em texto) e datas de criação e de confirmação do e-mail.',
        'Perfil: data de nascimento, sexo, altura, peso, cintura, percentual de gordura, nível de atividade e experiência com ciclismo.',
        'Saúde e limitações, que são dados pessoais sensíveis: lesões, dores, restrições médicas, cirurgias recentes e condições que afetam o exercício, quando você as informa.',
        'Contexto de ciclismo: rotina, equipamentos, metas, disponibilidade semanal e, se você informar, FTP e dados de potência.',
        'Treinos: planos gerados, sessões realizadas, esforço percebido (RPE), fadiga, dor, feedback, avaliações e check-ins de recuperação (sono, estresse e fadiga).',
        'Atividades importadas de arquivos .fit, .gpx ou .tcx que você envia: data e hora, duração, distância, altimetria e médias de frequência cardíaca, potência e cadência. O trajeto (coordenadas de GPS) é lido para calcular esses números e não é guardado.',
        'Mensagens que você envia pelo formulário de feedback do app.',
        'Dados técnicos: um cookie de sessão e registros de acesso do servidor e da rede, como endereço IP e horário, usados para segurança.',
      ],
    },
    {
      id: 'finalidades',
      title: 'Para que usamos e em que base legal',
      items: [
        'Montar e ajustar seus treinos, registrar sua evolução e mostrar os resultados: execução do serviço que você pediu (LGPD, art. 7º, V).',
        'Tratar seus dados de saúde e limitações para tornar os treinos mais seguros: seu consentimento específico, dado quando você informa esses dados (art. 11, I). Você pode retirá-lo a qualquer momento, apagando a informação no perfil ou encerrando a conta.',
        'Manter o serviço seguro, limitar tentativas de acesso indevido e investigar abusos: interesse legítimo (art. 7º, IX).',
        'Cumprir obrigações legais, quando houver.',
      ],
      after: [
        'Não usamos seus dados para publicidade, não os vendemos e não montamos perfil comercial.',
        'Os treinos são gerados por regras de planejamento definidas pelo produto, com base em estudos científicos citados no app, e não por decisão final de uma máquina sobre você: você decide se faz, cancela ou registra como não realizada cada sessão. Se quiser a revisão de um plano gerado de forma automatizada, peça pelo contato acima.',
        'O Cadência não faz diagnóstico nem tratamento médico.',
      ],
    },
    {
      id: 'compartilhamento',
      title: 'Com quem compartilhamos',
      paragraphs: [
        'Usamos prestadores de serviço (operadores) que tratam dados em nosso nome e só para o serviço combinado:',
      ],
      items: [
        'Hospedagem: servidor virtual em nuvem, onde ficam o aplicativo e o banco de dados, e a Cloudflare, que transporta o tráfego entre o seu aparelho e o servidor.',
        'E-mail transacional (Resend): recebe seu e-mail e o texto das mensagens de confirmação de e-mail e de redefinição de senha.',
        'Cópia de segurança: cópias criptografadas do banco de dados, em armazenamento externo.',
        'Explicação do treino por IA: quando você pede a explicação de uma sessão, um modelo de linguagem, acessado por um serviço na Cloudflare, recebe apenas o nome do treino, o objetivo, a duração, o esforço-alvo e as regras de planejamento aplicadas. Ele não recebe seu nome, e-mail, dados de saúde, histórico ou atividades importadas.',
      ],
      after: [
        'Cloudflare e Resend podem tratar dados fora do Brasil. Nesses casos, a transferência é necessária para prestar o serviço (LGPD, art. 33, IX).',
        'Não compartilhamos seus dados com anunciantes nem com outros usuários. Só os divulgamos a autoridades quando a lei exigir.',
      ],
    },
    {
      id: 'retencao',
      title: 'Por quanto tempo guardamos',
      items: [
        'Seus dados ficam guardados enquanto a conta existir.',
        'Ao encerrar a conta em Configurações, apagamos todos os seus dados do banco de dados na hora, sem intervenção de ninguém.',
        'Cópias de segurança anteriores ao encerramento permanecem por um prazo limitado e são descartadas: no servidor, por até 14 dias. As cópias externas são criptografadas e seguem a regra de retenção do armazenamento.',
        'Registros técnicos de acesso ficam pelo tempo necessário para segurança.',
      ],
    },
    {
      id: 'direitos',
      title: 'Seus direitos',
      paragraphs: ['Você pode, a qualquer momento e sem custo (LGPD, art. 18):'],
      items: [
        'Confirmar que tratamos seus dados e acessá-los. Em Configurações, o botão "Exportar meus dados" baixa uma cópia completa em JSON.',
        'Corrigir dados incompletos ou desatualizados: em Perfil, e nos registros de treino que permitem correção.',
        'Levar seus dados a outro serviço: o arquivo exportado é estruturado e legível por máquina.',
        'Apagar seus dados: em Configurações, "Encerrar conta".',
        'Saber com quem compartilhamos dados, conforme a seção acima.',
        'Retirar o consentimento e pedir a revisão de decisões automatizadas.',
      ],
      after: [
        `Para o que não puder fazer sozinho, escreva para ${LEGAL_CONTACT_EMAIL}. Respondemos em até 15 dias.`,
        'Você também pode reclamar à Autoridade Nacional de Proteção de Dados (ANPD), em gov.br/anpd.',
      ],
    },
    {
      id: 'seguranca',
      title: 'Como protegemos seus dados',
      items: [
        'Conexão criptografada (HTTPS) entre o seu aparelho e o servidor.',
        'Senhas guardadas com hash bcrypt; o cookie de sessão é HttpOnly, seguro e protegido contra envio entre sites.',
        'Cada conta só acessa os próprios dados; você pode encerrar as sessões abertas em outros aparelhos.',
        'Limite de tentativas em login, cadastro e redefinição de senha.',
        'Cópias de segurança criptografadas.',
      ],
      after: [
        'Nenhum sistema é totalmente seguro. Se ocorrer um incidente que possa causar risco ou dano relevante a você, avisaremos você e a ANPD, como a lei determina (art. 48).',
      ],
    },
    {
      id: 'cookies',
      title: 'Cookies e armazenamento no aparelho',
      paragraphs: [
        'Usamos um cookie essencial, que mantém você conectado, e guardamos no navegador apenas preferências da interface, como saber se o app foi instalado. Não usamos cookies de publicidade nem ferramentas de análise de terceiros.',
      ],
    },
    {
      id: 'menores',
      title: 'Crianças e adolescentes',
      paragraphs: [
        'O Cadência é destinado a pessoas com 18 anos ou mais. Se identificarmos uma conta de menor, apagaremos os dados.',
      ],
    },
    {
      id: 'mudancas',
      title: 'Mudanças nesta política',
      paragraphs: [
        'Quando mudarmos algo que afete seus dados, como um novo provedor ou uma nova integração, atualizamos esta página e a data no topo, e avisamos no app antes de a mudança valer. Uma integração nova que exija novos dados só começa com o seu consentimento.',
      ],
    },
  ],
};

export const TERMS_OF_USE: LegalDocument = {
  kicker: 'TERMOS',
  title: 'Termos de Uso',
  intro:
    'Estes termos valem para o uso do Cadência, um aplicativo de planejamento de treino de ciclismo. Ao criar uma conta, você declara que leu e concorda com eles e com a Política de Privacidade.',
  sections: [
    {
      id: 'elegibilidade',
      title: 'Quem pode usar',
      paragraphs: [
        'O Cadência é destinado a pessoas com 18 anos ou mais. A conta é pessoal e você deve informar dados verdadeiros.',
      ],
    },
    {
      id: 'servico',
      title: 'O que o Cadência é, e o que não é',
      paragraphs: [
        'O Cadência sugere treinos de ciclismo a partir do seu perfil, da sua rotina e dos seus registros. Ele não é serviço médico, de fisioterapia ou de nutrição, não faz diagnóstico e não substitui a orientação de profissionais de saúde ou de educação física.',
        'Antes de começar ou de aumentar a carga de treino, principalmente se você tem doença, lesão, restrição médica ou cirurgia recente, consulte um profissional. Interrompa o treino diante de dor, tontura, falta de ar ou qualquer mal-estar. Você decide se faz cada treino e como o faz, e responde por essa decisão.',
      ],
    },
    {
      id: 'conta',
      title: 'Sua conta',
      items: [
        'Guarde sua senha e não a compartilhe. Você é responsável pelo que acontece na sua conta.',
        'Avise se suspeitar de acesso indevido. Em Configurações, você pode trocar a senha e encerrar as sessões abertas em outros aparelhos.',
        'Você pode encerrar a conta a qualquer momento, em Configurações, e seus dados são apagados.',
      ],
    },
    {
      id: 'uso-adequado',
      title: 'Uso adequado',
      paragraphs: ['Não é permitido:'],
      items: [
        'tentar acessar dados de outras contas ou contornar a segurança do app;',
        'sobrecarregar o serviço ou usar meios automatizados que prejudiquem outros usuários;',
        'enviar arquivos ou textos ilegais, ofensivos ou que não sejam seus.',
      ],
      after: ['Podemos suspender contas que violem estas regras.'],
    },
    {
      id: 'seus-dados',
      title: 'Seus dados e conteúdo',
      paragraphs: [
        'Os dados e arquivos que você envia continuam sendo seus. Você nos autoriza a tratá-los só para operar o Cadência para você, como descrito na Política de Privacidade.',
      ],
    },
    {
      id: 'ia-e-ciencia',
      title: 'Explicações por IA e base científica',
      paragraphs: [
        'Os treinos vêm de regras de planejamento. As explicações escritas por inteligência artificial resumem essas regras e podem conter imprecisões: confie no treino e nas regras mostradas, não no texto de uma IA.',
        'Os estudos citados no app dão base geral a princípios de treinamento. Limiares, prazos e percentuais do app são escolhas de produto, e o app informa quando é o caso. Estudos de grupos populacionais não garantem o mesmo resultado para você.',
      ],
    },
    {
      id: 'disponibilidade',
      title: 'Disponibilidade e mudanças',
      paragraphs: [
        'O Cadência está em evolução e é oferecido como está. Pode ficar indisponível para manutenção ou por falha, e recursos podem mudar ou sair. Faremos o possível para avisar antes de mudanças que afetem seus dados.',
      ],
    },
    {
      id: 'responsabilidade',
      title: 'Responsabilidade',
      paragraphs: [
        'Dentro do que a lei permite, o Cadência não garante resultados de desempenho, de saúde ou de composição corporal, e não responde por lesões, danos ou perdas decorrentes do treino que você decidiu fazer. Nada aqui afasta direitos que a lei assegura a você, inclusive os do Código de Defesa do Consumidor.',
        'Integrações com serviços de terceiros, quando existirem, são oferecidas por esses terceiros. Eles não dão garantias nem assumem responsabilidade sobre o Cadência.',
      ],
    },
    {
      id: 'lei',
      title: 'Lei aplicável',
      paragraphs: [
        'Valem as leis do Brasil. Se você for consumidor, o foro é o do seu domicílio.',
      ],
    },
    {
      id: 'mudancas',
      title: 'Mudanças e contato',
      paragraphs: [
        `Podemos atualizar estes termos; a data no topo mostra a versão em vigor, e avisamos no app quando a mudança for relevante. Dúvidas: ${LEGAL_CONTACT_EMAIL}.`,
      ],
    },
  ],
};
