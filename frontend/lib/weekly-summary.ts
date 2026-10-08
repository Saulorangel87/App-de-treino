// Resumo semanal por regras: lê os números que a tela de Evolução já recebe e os
// transforma em poucas frases sobre a última semana completa (segunda a domingo).
// Não usa IA nem envia nada para fora; o mesmo dado gera sempre o mesmo texto.
// As frases são do Cadência, nos dois idiomas, e não são diagnóstico.

import { formatDecimal, INTL_LOCALE, type Locale } from './i18n';
import { zoneLabel } from './zones';

type WeekInput = {
  week_start: string;
  completed_sessions: number;
  cancelled_sessions: number;
  total_minutes: number;
  average_rpe: number;
};

type SessionInput = {
  completed_on: string;
  rpe_delta?: number;
  fatigue_after?: number;
  pain_reported: boolean;
};

type RecoveryInput = { recorded_on: string; fatigue_level: number };

export type WeeklySummaryInput = {
  weeks: WeekInput[];
  recent_sessions?: SessionInput[];
  recovery: RecoveryInput[];
};

export type SummaryTone = 'good' | 'neutral' | 'warn';
export type SummaryLine = { text: string; tone: SummaryTone };
export type WeeklySummary = { title: string; range: string; lines: SummaryLine[]; footnote: string };

const text = {
  pt: {
    title: 'Sua semana',
    footnote: 'Resumo montado pelas regras do app a partir dos seus registros. Não é diagnóstico nem substitui orientação profissional.',
    sessions: (done: number, cancelled: number) =>
      `Você concluiu ${done} ${done === 1 ? 'treino' : 'treinos'}${cancelled > 0 ? ` e cancelou ${cancelled}` : ''}.`,
    timeSame: (minutes: string) => `Foram ${minutes} de pedal, parecido com a semana anterior.`,
    timeMore: (minutes: string, percent: number) => `Foram ${minutes} de pedal, ${percent}% a mais que na semana anterior.`,
    timeLess: (minutes: string, percent: number) => `Foram ${minutes} de pedal, ${percent}% a menos que na semana anterior.`,
    timeOnly: (minutes: string) => `Foram ${minutes} de pedal.`,
    zone: (zone: string) => `A zona média registrada foi ${zone}.`,
    above: (count: number) =>
      `${count} ${count === 1 ? 'treino ficou' : 'treinos ficaram'} com esforço acima do previsto.`,
    fatigue: (value: string) => `A fadiga média foi ${value} de 5.`,
    fatigueHigh: 'É alta: vale priorizar descanso e sono.',
    painNone: 'Nenhuma dor foi relatada.',
    pain: 'Você relatou dor. Se ela persistir, procure um profissional de saúde.',
    nextCare: 'O Cadência leva isso em conta ao ajustar os próximos treinos.',
    nextOk: 'Nada nesta semana pede cuidado extra.',
  },
  en: {
    title: 'Your week',
    footnote: "Summary built by the app's rules from your records. It is not a diagnosis and does not replace professional guidance.",
    sessions: (done: number, cancelled: number) =>
      `You completed ${done} ${done === 1 ? 'workout' : 'workouts'}${cancelled > 0 ? ` and cancelled ${cancelled}` : ''}.`,
    timeSame: (minutes: string) => `That was ${minutes} of riding, similar to the previous week.`,
    timeMore: (minutes: string, percent: number) => `That was ${minutes} of riding, ${percent}% more than the previous week.`,
    timeLess: (minutes: string, percent: number) => `That was ${minutes} of riding, ${percent}% less than the previous week.`,
    timeOnly: (minutes: string) => `That was ${minutes} of riding.`,
    zone: (zone: string) => `The average zone recorded was ${zone}.`,
    above: (count: number) => `${count} ${count === 1 ? 'workout' : 'workouts'} felt harder than planned.`,
    fatigue: (value: string) => `Average fatigue was ${value} out of 5.`,
    fatigueHigh: "That's high: it's worth prioritizing rest and sleep.",
    painNone: 'No pain was reported.',
    pain: 'You reported pain. If it persists, see a health professional.',
    nextCare: 'Cadência takes this into account when adjusting your next workouts.',
    nextOk: 'Nothing this week calls for extra care.',
  },
};

const DAY_MS = 24 * 60 * 60 * 1000;

function parseDay(value: string): Date {
  return new Date(`${value}T12:00:00`);
}

function dayKey(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${date.getFullYear()}-${month}-${day}`;
}

function addDays(value: string, days: number): string {
  return dayKey(new Date(parseDay(value).getTime() + days * DAY_MS));
}

/** Segunda-feira da semana de uma data local (AAAA-MM-DD). */
export function mondayOf(date: Date): string {
  const noon = new Date(date.getFullYear(), date.getMonth(), date.getDate(), 12);
  return dayKey(new Date(noon.getTime() - ((noon.getDay() + 6) % 7) * DAY_MS));
}

function formatMinutes(total: number): string {
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  if (!hours) return `${minutes} min`;
  return minutes ? `${hours}h${String(minutes).padStart(2, '0')}` : `${hours}h`;
}

function inWeek(day: string, start: string): boolean {
  return day >= start && day <= addDays(start, 6);
}

/**
 * Resumo da última semana completa (a anterior à semana de `today`). Devolve
 * null quando essa semana não tem treino concluído nem cancelado.
 */
export function buildWeeklySummary(input: WeeklySummaryInput, locale: Locale, today: Date): WeeklySummary | null {
  const start = addDays(mondayOf(today), -7);
  const week = input.weeks.find((item) => item.week_start === start);
  if (!week || week.completed_sessions + week.cancelled_sessions === 0) return null;
  const previous = input.weeks.find((item) => item.week_start === addDays(start, -7));
  const t = text[locale];

  const sessions = (input.recent_sessions ?? []).filter((item) => inWeek(item.completed_on, start));
  const checkins = input.recovery.filter((item) => inWeek(item.recorded_on, start));
  const lines: SummaryLine[] = [];

  lines.push({
    text: t.sessions(week.completed_sessions, week.cancelled_sessions),
    tone: week.cancelled_sessions > 0 ? 'neutral' : 'good',
  });

  if (week.total_minutes > 0) {
    const minutes = formatMinutes(week.total_minutes);
    const change = previous && previous.total_minutes > 0 ? ((week.total_minutes - previous.total_minutes) / previous.total_minutes) * 100 : null;
    const percent = change === null ? 0 : Math.round(Math.abs(change));
    const sentence =
      change === null ? t.timeOnly(minutes) : change >= 5 ? t.timeMore(minutes, percent) : change <= -5 ? t.timeLess(minutes, percent) : t.timeSame(minutes);
    lines.push({ text: sentence, tone: 'neutral' });
  }

  if (week.average_rpe > 0) lines.push({ text: t.zone(zoneLabel(week.average_rpe, locale)), tone: 'neutral' });

  const above = sessions.filter((item) => (item.rpe_delta ?? 0) >= 2).length;
  if (above > 0) lines.push({ text: t.above(above), tone: 'warn' });

  const fatigueValues = [
    ...sessions.filter((item) => item.fatigue_after !== undefined).map((item) => item.fatigue_after as number),
    ...checkins.map((item) => item.fatigue_level),
  ];
  const fatigue = fatigueValues.length ? fatigueValues.reduce((sum, value) => sum + value, 0) / fatigueValues.length : null;
  const fatigueHigh = fatigue !== null && fatigue >= 4;
  if (fatigue !== null) {
    const sentence = t.fatigue(formatDecimal(fatigue, 1, locale));
    lines.push({ text: fatigueHigh ? `${sentence} ${t.fatigueHigh}` : sentence, tone: fatigueHigh ? 'warn' : 'neutral' });
  }

  const pain = sessions.some((item) => item.pain_reported);
  lines.push({ text: pain ? t.pain : t.painNone, tone: pain ? 'warn' : 'good' });

  const care = pain || fatigueHigh || above > 0;
  lines.push({ text: care ? t.nextCare : t.nextOk, tone: care ? 'warn' : 'good' });

  const format = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: 'numeric', month: 'short' });
  const clean = (value: string) => format.format(parseDay(value)).replace(/\./g, '');
  return { title: t.title, range: `${clean(start)} – ${clean(addDays(start, 6))}`, lines, footnote: t.footnote };
}
