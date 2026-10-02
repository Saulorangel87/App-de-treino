// Zonas de esforço (Z1 a Z5): a linguagem do app para a intensidade do treino. O
// motor continua prescrevendo e adaptando em RPE; as zonas e o RPE se ligam por
// esta tabela. Os limites entre zonas são escolha do Cadência, não resultado de
// estudo, e precisam ser iguais aos de backend/internal/zones/zones.go (um teste
// do backend confere o número, o nome, maxRpe e feedbackRpe de cada zona).

export type Zone = {
  number: 1 | 2 | 3 | 4 | 5;
  name: string;
  /** Maior RPE prescrito que ainda pertence à zona (a última não tem teto). */
  maxRpe: number;
  /** RPE gravado quando o atleta diz que pedalou nesta zona. */
  feedbackRpe: number;
  /** Teste da conversa: o que dá para fazer enquanto pedala. */
  talk: string;
  /** Faixa em % da frequência cardíaca máxima [de, até]. */
  hrPercent: [number, number];
  /** Faixa em % do FTP; null no limite aberto. */
  ftpPercent: [number | null, number | null];
};

export const ZONES: readonly Zone[] = [
  {
    number: 1, name: 'Recuperação', maxRpe: 3.5, feedbackRpe: 3,
    talk: 'Conversa fácil, respiração tranquila e pernas leves.',
    hrPercent: [50, 60], ftpPercent: [null, 55],
  },
  {
    number: 2, name: 'Resistência', maxRpe: 5, feedbackRpe: 4.5,
    talk: 'Dá para conversar em frases completas. Esforço sustentável por horas.',
    hrPercent: [60, 70], ftpPercent: [55, 75],
  },
  {
    number: 3, name: 'Ritmo', maxRpe: 6.5, feedbackRpe: 6,
    talk: 'Só frases curtas. Esforço firme, mas controlado.',
    hrPercent: [70, 80], ftpPercent: [75, 90],
  },
  {
    number: 4, name: 'Limiar', maxRpe: 7.5, feedbackRpe: 7,
    talk: 'Poucas palavras por vez. Desconforto forte, sustentável por 20 a 40 minutos.',
    hrPercent: [80, 90], ftpPercent: [90, 105],
  },
  {
    number: 5, name: 'Intenso', maxRpe: 10, feedbackRpe: 8.5,
    talk: 'Não dá para conversar. Esforço de poucos minutos, com recuperação entre eles.',
    hrPercent: [90, 100], ftpPercent: [105, null],
  },
];

export const ZONE_DISCLAIMER =
  'Os limites entre as zonas são critérios do Cadência, e as faixas em batimentos e watts são referências: variam de pessoa para pessoa e com calor, cansaço e cafeína.';

export type ZoneReference = {
  maxHeartRate?: number;
  ftp?: number;
};

export type ZoneRanges = {
  heartRate?: { low: number; high: number };
  power?: { low?: number; high?: number };
};

export function zoneForRpe(rpe: number): Zone {
  return ZONES.find((zone) => rpe <= zone.maxRpe) ?? ZONES[ZONES.length - 1];
}

export function rpeForZone(number: number): number | undefined {
  return ZONES.find((zone) => zone.number === number)?.feedbackRpe;
}

/** Texto curto de uma zona, como "Z2 · Resistência". */
export function zoneLabel(rpe: number): string {
  const zone = zoneForRpe(rpe);
  return `Z${zone.number} · ${zone.name}`;
}

/**
 * Faixas da zona em batimentos e em watts, quando o atleta informou a frequência
 * cardíaca máxima e/ou o FTP. As zonas vizinhas não se sobrepõem: a faixa de uma
 * começa um batimento (ou watt) acima do fim da anterior.
 */
export function zoneRanges(zone: Zone, reference: ZoneReference): ZoneRanges {
  const ranges: ZoneRanges = {};
  const { maxHeartRate, ftp } = reference;
  if (maxHeartRate && maxHeartRate > 0) {
    const edge = (percent: number) => Math.round((maxHeartRate * percent) / 100);
    const [from, to] = zone.hrPercent;
    ranges.heartRate = { low: zone.number === 1 ? edge(from) : edge(from) + 1, high: edge(to) };
  }
  if (ftp && ftp > 0) {
    const edge = (percent: number) => Math.round((ftp * percent) / 100);
    const [from, to] = zone.ftpPercent;
    ranges.power = {
      low: from === null ? undefined : edge(from) + 1,
      high: to === null ? undefined : edge(to),
    };
  }
  return ranges;
}

export function formatHeartRateRange(range: NonNullable<ZoneRanges['heartRate']>): string {
  return `${range.low}–${range.high} bpm`;
}

export function formatPowerRange(range: NonNullable<ZoneRanges['power']>): string {
  if (range.low === undefined && range.high !== undefined) return `até ${range.high} W`;
  if (range.high === undefined && range.low !== undefined) return `acima de ${range.low - 1} W`;
  return `${range.low}–${range.high} W`;
}

/** Referências do atleta para as faixas, lidas do contexto de ciclismo do plano. */
export function zoneReferenceFrom(cyclingContext?: {
  max_heart_rate?: number;
  ftp?: number;
  uses_heart_rate?: boolean;
  uses_power?: boolean;
}): ZoneReference {
  if (!cyclingContext) return {};
  return {
    maxHeartRate: cyclingContext.uses_heart_rate ? cyclingContext.max_heart_rate : undefined,
    ftp: cyclingContext.uses_power ? cyclingContext.ftp : undefined,
  };
}
