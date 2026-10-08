import { describe, expect, it } from 'vitest';
import {
  activeProtection,
  addDays,
  durationSourceLabel,
  formatTrainingDay,
  isFutureTrainingDate,
  loggableDateRange,
  undoConfirmation,
  type Workout,
} from './planning';

function workout(status: string, protection?: Workout['explanation']['protection']): Pick<Workout, 'status' | 'explanation'> {
  return { status, explanation: { protection } } as Pick<Workout, 'status' | 'explanation'>;
}

describe('activeProtection', () => {
  it('ignora planos sem bloco de proteção (gerados antes da proteção graduada)', () => {
    expect(activeProtection({ workouts: [workout('planned')] as Workout[] })).toBeNull();
  });

  it('ignora o nível "none" e treinos que não são mais planejados', () => {
    const strong = { level: 'strong' as const, reasons: ['dor'], suggest_professional: false };
    const none = { level: 'none' as const, reasons: [], suggest_professional: false };
    expect(activeProtection({ workouts: [workout('planned', none), workout('completed', strong)] as Workout[] })).toBeNull();
  });

  it('devolve a proteção mais forte entre os treinos planejados', () => {
    const light = { level: 'light' as const, reasons: ['fadiga'], suggest_professional: false };
    const strong = { level: 'strong' as const, reasons: ['dor'], suggest_professional: true, expires_on: '2026-10-07' };
    expect(activeProtection({ workouts: [workout('planned', light), workout('planned', strong)] as Workout[] })).toBe(strong);
  });
});

describe('isFutureTrainingDate', () => {
  it('bloqueia só datas depois de hoje', () => {
    expect(isFutureTrainingDate('2026-10-03', '2026-10-02')).toBe(true);
    expect(isFutureTrainingDate('2026-10-02', '2026-10-02')).toBe(false);
    expect(isFutureTrainingDate('2026-10-01', '2026-10-02')).toBe(false);
  });

  it('não decide antes de conhecer a data local', () => {
    expect(isFutureTrainingDate('2026-10-03', '')).toBe(false);
  });
});

describe('formatTrainingDay', () => {
  it('mostra dia e mês', () => {
    expect(formatTrainingDay('2026-10-03', 'pt')).toBe('03/10');
    expect(formatTrainingDay('2026-10-03', 'en')).toBe('10/03');
    expect(formatTrainingDay('invalida', 'pt')).toBe('invalida');
  });
});

describe('undoConfirmation', () => {
  it('explica o que será revertido ao desfazer um treino concluído', () => {
    const text = undoConfirmation('completed', 'pt');
    expect(text).toContain('sessão e o feedback serão apagados');
    expect(text).toContain('ajustes');
    expect(undoConfirmation('completed', 'en')).toContain('cannot be undone');
  });

  it('é mais simples ao reabrir um treino não realizado', () => {
    expect(undoConfirmation('skipped', 'pt')).toContain('Reabrir');
    expect(undoConfirmation('skipped', 'en')).toContain('Reopen');
  });
});

describe('addDays', () => {
  it('soma e subtrai dias atravessando mês e ano', () => {
    expect(addDays('2026-10-01', -7)).toBe('2026-09-24');
    expect(addDays('2026-01-03', -7)).toBe('2025-12-27');
    expect(addDays('2026-02-27', 3)).toBe('2026-03-02');
  });
});

describe('datas ainda desconhecidas', () => {
  it('não quebram na primeira renderização, antes de saber a data local', () => {
    expect(addDays('', -7)).toBe('');
    expect(addDays('abc', 1)).toBe('abc');
    expect(loggableDateRange('2026-10-01', '')).toEqual({ min: '', max: '' });
  });
});

describe('loggableDateRange', () => {
  it('permite até 7 dias atrás e hoje', () => {
    expect(loggableDateRange('2026-09-01', '2026-10-01')).toEqual({ min: '2026-09-24', max: '2026-10-01' });
  });

  it('não deixa registrar antes do dia planejado', () => {
    expect(loggableDateRange('2026-09-30', '2026-10-01')).toEqual({ min: '2026-09-30', max: '2026-10-01' });
    expect(loggableDateRange('2026-10-01', '2026-10-01')).toEqual({ min: '2026-10-01', max: '2026-10-01' });
  });
});

describe('durationSourceLabel', () => {
  it('diz de onde vem a duração', () => {
    expect(durationSourceLabel('reported', 'pt')).toBe('informada por você');
    expect(durationSourceLabel('imported', 'pt')).toBe('de um arquivo importado');
    expect(durationSourceLabel('timer', 'pt')).toBe('medida pelo cronômetro');
    expect(durationSourceLabel(undefined, 'pt')).toBe('medida pelo cronômetro');
    expect(durationSourceLabel('reported', 'en')).toBe('entered by you');
    expect(durationSourceLabel(undefined, 'en')).toBe('measured by the timer');
  });
});
