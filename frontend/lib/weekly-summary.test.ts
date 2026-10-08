import { describe, expect, it } from 'vitest';
import { buildWeeklySummary, mondayOf, type WeeklySummaryInput } from './weekly-summary';

// Hoje: quarta, 7/10/2026. A última semana completa é a de 28/09 a 04/10; a anterior, 21/09.
const today = new Date(2026, 9, 7, 10);

function week(start: string, patch: Partial<WeeklySummaryInput['weeks'][number]> = {}) {
  return { week_start: start, completed_sessions: 0, cancelled_sessions: 0, total_minutes: 0, average_rpe: 0, ...patch };
}

function input(patch: Partial<WeeklySummaryInput> = {}): WeeklySummaryInput {
  return {
    weeks: [
      week('2026-09-21', { completed_sessions: 2, total_minutes: 100 }),
      week('2026-09-28', { completed_sessions: 3, total_minutes: 150, average_rpe: 4.5 }),
    ],
    recent_sessions: [],
    recovery: [],
    ...patch,
  };
}

describe('mondayOf', () => {
  it('acha a segunda-feira de qualquer dia da semana', () => {
    expect(mondayOf(new Date(2026, 9, 7))).toBe('2026-10-05');
    expect(mondayOf(new Date(2026, 9, 5))).toBe('2026-10-05');
    expect(mondayOf(new Date(2026, 9, 11))).toBe('2026-10-05');
    expect(mondayOf(new Date(2026, 0, 1))).toBe('2025-12-29');
  });
});

describe('buildWeeklySummary', () => {
  it('resume a última semana completa, com a comparação com a anterior', () => {
    const summary = buildWeeklySummary(input(), 'pt', today)!;
    const texts = summary.lines.map((line) => line.text);
    expect(summary.range).toBe('28 de set – 4 de out');
    expect(texts[0]).toBe('Você concluiu 3 treinos.');
    expect(texts[1]).toBe('Foram 2h30 de pedal, 50% a mais que na semana anterior.');
    expect(texts).toContain('A zona média registrada foi Z2 · Resistência.');
    expect(texts).toContain('Nenhuma dor foi relatada.');
    expect(texts[texts.length - 1]).toBe('Nada nesta semana pede cuidado extra.');
  });

  it('não cria resumo quando a semana passada não tem treino', () => {
    expect(buildWeeklySummary(input({ weeks: [week('2026-09-28')] }), 'pt', today)).toBeNull();
    expect(buildWeeklySummary(input({ weeks: [week('2026-10-05', { completed_sessions: 1 })] }), 'pt', today)).toBeNull();
  });

  it('conta treinos cancelados e trata o singular', () => {
    const summary = buildWeeklySummary(input({ weeks: [week('2026-09-28', { completed_sessions: 1, cancelled_sessions: 2, total_minutes: 40 })] }), 'pt', today)!;
    expect(summary.lines[0]).toEqual({ text: 'Você concluiu 1 treino e cancelou 2.', tone: 'neutral' });
    expect(summary.lines[1].text).toBe('Foram 40 min de pedal.');
  });

  it('chama atenção para dor, esforço acima do previsto e fadiga alta, só dentro da semana', () => {
    const summary = buildWeeklySummary(
      input({
        recent_sessions: [
          { completed_on: '2026-09-30', rpe_delta: 3, fatigue_after: 5, pain_reported: true },
          { completed_on: '2026-10-02', rpe_delta: 0, fatigue_after: 4, pain_reported: false },
          { completed_on: '2026-10-06', rpe_delta: 4, fatigue_after: 1, pain_reported: true },
        ],
      }),
      'pt',
      today,
    )!;
    const texts = summary.lines.map((line) => line.text);
    expect(texts).toContain('1 treino ficou com esforço acima do previsto.');
    expect(texts).toContain('A fadiga média foi 4,5 de 5. É alta: vale priorizar descanso e sono.');
    expect(texts).toContain('Você relatou dor. Se ela persistir, procure um profissional de saúde.');
    expect(texts[texts.length - 1]).toBe('O Cadência leva isso em conta ao ajustar os próximos treinos.');
    expect(summary.lines.at(-1)?.tone).toBe('warn');
  });

  it('usa os check-ins na fadiga e não inventa fadiga sem dado', () => {
    const withCheckins = buildWeeklySummary(input({ recovery: [{ recorded_on: '2026-10-01', fatigue_level: 2 }, { recorded_on: '2026-10-20', fatigue_level: 5 }] }), 'pt', today)!;
    expect(withCheckins.lines.map((line) => line.text)).toContain('A fadiga média foi 2,0 de 5.');
    const none = buildWeeklySummary(input(), 'pt', today)!;
    expect(none.lines.some((line) => line.text.includes('fadiga'))).toBe(false);
  });

  it('fica parecido quando a variação é pequena e diz menos quando cai', () => {
    const similar = buildWeeklySummary(input({ weeks: [week('2026-09-21', { completed_sessions: 1, total_minutes: 100 }), week('2026-09-28', { completed_sessions: 1, total_minutes: 102 })] }), 'pt', today)!;
    expect(similar.lines[1].text).toBe('Foram 1h42 de pedal, parecido com a semana anterior.');
    const less = buildWeeklySummary(input({ weeks: [week('2026-09-21', { completed_sessions: 1, total_minutes: 100 }), week('2026-09-28', { completed_sessions: 1, total_minutes: 60 })] }), 'pt', today)!;
    expect(less.lines[1].text).toBe('Foram 1h de pedal, 40% a menos que na semana anterior.');
  });

  it('escreve em inglês, com números no formato do idioma', () => {
    const summary = buildWeeklySummary(
      input({ recent_sessions: [{ completed_on: '2026-09-30', rpe_delta: 2, fatigue_after: 3, pain_reported: false }] }),
      'en',
      today,
    )!;
    const texts = summary.lines.map((line) => line.text);
    expect(summary.title).toBe('Your week');
    expect(summary.range).toBe('Sep 28 – Oct 4');
    expect(texts[0]).toBe('You completed 3 workouts.');
    expect(texts[1]).toBe('That was 2h30 of riding, 50% more than the previous week.');
    expect(texts).toContain('1 workout felt harder than planned.');
    expect(texts).toContain('Average fatigue was 3.0 out of 5.');
    expect(texts).toContain('The average zone recorded was Z2 · Endurance.');
    expect(texts.at(-1)).toBe('Cadência takes this into account when adjusting your next workouts.');
  });
});
