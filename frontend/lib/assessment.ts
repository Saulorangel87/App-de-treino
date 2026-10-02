// Avaliação: o pedal de referência em Z2 e o que a tela mostra dele. Os números
// derivados (eficiência e deriva de frequência cardíaca) vêm do servidor; aqui
// ficam a leitura deles, a comparação com a avaliação anterior e o prazo para refazer.

export type AssessmentEfficiency = { value: number; kind: 'power_per_bpm' | 'speed_per_100bpm' };

export type Assessment = {
  id: string;
  completed_at?: string;
  duration_minutes: number;
  actual_rpe: number;
  pain_reported: boolean;
  eligible_for_progression: boolean;
  notes?: string;
  average_heart_rate?: number;
  average_power_w?: number;
  distance_km?: number;
  heart_rate_first_half?: number;
  heart_rate_second_half?: number;
  heart_rate_drift_percent?: number;
  efficiency?: AssessmentEfficiency;
};

/** Prazo a partir do qual o app sugere refazer; a faixa dita ao atleta é de 4 a 6 semanas. */
export const REASSESS_AFTER_DAYS = 28;
export const REASSESS_RANGE_TEXT = '4 a 6 semanas';

const DAY_MS = 24 * 60 * 60 * 1000;

export function efficiencyUnit(kind: AssessmentEfficiency['kind']): string {
  return kind === 'power_per_bpm' ? 'W por bpm' : 'km/h a cada 100 bpm';
}

export function formatEfficiency(efficiency: AssessmentEfficiency): string {
  return `${efficiency.value.toFixed(2).replace('.', ',')} ${efficiencyUnit(efficiency.kind)}`;
}

/**
 * Variação percentual da eficiência em relação à avaliação anterior. Só compara
 * valores do mesmo tipo (potência com potência, velocidade com velocidade).
 */
export function efficiencyChange(
  current?: AssessmentEfficiency,
  previous?: AssessmentEfficiency,
): number | null {
  if (!current || !previous || current.kind !== previous.kind || previous.value <= 0) return null;
  return Math.round(((current.value - previous.value) / previous.value) * 1000) / 10;
}

export function formatChange(percent: number): string {
  const text = Math.abs(percent).toFixed(1).replace('.', ',');
  if (percent > 0) return `${text}% melhor`;
  if (percent < 0) return `${text}% pior`;
  return 'igual';
}

export type DriftReading = { tone: 'steady' | 'rising' | 'high'; text: string };

/**
 * Leitura da deriva: quanto a frequência cardíaca subiu da primeira para a
 * segunda metade do pedal, com a mesma potência ou o mesmo ritmo. Até 5% é o
 * esperado para um pedal em Z2; acima disso o ritmo pode ter ficado forte para
 * a zona, ou houve calor, pouca hidratação ou cansaço. É informativo: não muda o "apto".
 */
export function describeDrift(percent: number): DriftReading {
  if (percent <= 3) {
    return { tone: 'steady', text: 'Frequência cardíaca estável: o ritmo combinou com a zona.' };
  }
  if (percent <= 5) {
    return { tone: 'rising', text: 'Subiu um pouco, dentro do esperado para um pedal em Z2.' };
  }
  return {
    tone: 'high',
    text: 'Subiu bastante: o ritmo pode ter ficado forte para Z2, ou houve calor, pouca água ou cansaço.',
  };
}

export function reassessmentDate(completedAt: string): Date {
  return new Date(new Date(completedAt).getTime() + REASSESS_AFTER_DAYS * DAY_MS);
}

export type ReassessmentStatus = { due: boolean; daysLeft: number };

export function reassessmentStatus(completedAt: string, now: Date): ReassessmentStatus {
  const daysLeft = Math.ceil((reassessmentDate(completedAt).getTime() - now.getTime()) / DAY_MS);
  return { due: daysLeft <= 0, daysLeft: Math.max(daysLeft, 0) };
}

export type ImportedRide = {
  moving_seconds: number;
  distance_km: number;
  average_heart_rate?: number;
  average_power_watts?: number;
};

/** O pedal importado só serve se couber na duração aceita pela avaliação (15 a 30 minutos). */
export function fitsAssessment(ride: Pick<ImportedRide, 'moving_seconds'>): boolean {
  const minutes = Math.round(ride.moving_seconds / 60);
  return minutes >= 15 && minutes <= 30;
}

export type AssessmentFields = {
  duration: string;
  averageHeartRate: string;
  averagePower: string;
  distanceKm: string;
};

/** Preenche o formulário com os números de uma atividade importada (campos vazios quando ela não os tem). */
export function fieldsFromRide(ride: ImportedRide): AssessmentFields {
  return {
    duration: String(Math.round(ride.moving_seconds / 60)),
    averageHeartRate: ride.average_heart_rate ? String(Math.round(ride.average_heart_rate)) : '',
    averagePower: ride.average_power_watts ? String(Math.round(ride.average_power_watts)) : '',
    distanceKm: ride.distance_km > 0 ? String(Math.round(ride.distance_km * 100) / 100) : '',
  };
}

function optionalInteger(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text.replace(',', '.'));
  return Number.isFinite(value) ? Math.round(value) : undefined;
}

function optionalDecimal(raw: string): number | undefined {
  const text = raw.trim();
  if (!text) return undefined;
  const value = Number(text.replace(',', '.'));
  return Number.isFinite(value) ? value : undefined;
}

export type AssessmentNumbersInput = {
  averageHeartRate: string;
  averagePower: string;
  distanceKm: string;
  firstHalf: string;
  secondHalf: string;
};

/** Campos de texto do formulário para o corpo da requisição; vazio vira "não informado". */
export function numbersPayload(input: AssessmentNumbersInput) {
  return {
    average_heart_rate: optionalInteger(input.averageHeartRate),
    average_power_w: optionalInteger(input.averagePower),
    distance_km: optionalDecimal(input.distanceKm),
    heart_rate_first_half: optionalInteger(input.firstHalf),
    heart_rate_second_half: optionalInteger(input.secondHalf),
  };
}

/** As duas metades da frequência cardíaca andam juntas: ou as duas, ou nenhuma. */
export function halvesError(firstHalf: string, secondHalf: string): string {
  const first = firstHalf.trim() !== '';
  const second = secondHalf.trim() !== '';
  return first === second ? '' : 'Informe a frequência cardíaca das duas metades, ou deixe as duas em branco.';
}
