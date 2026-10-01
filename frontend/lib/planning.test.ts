import { describe, expect, it } from 'vitest';
import {
  activeProtection,
  formatTrainingDay,
  isFutureTrainingDate,
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
    expect(formatTrainingDay('2026-10-03')).toBe('03/10');
    expect(formatTrainingDay('invalida')).toBe('invalida');
  });
});

describe('undoConfirmation', () => {
  it('explica o que será revertido ao desfazer um treino concluído', () => {
    const text = undoConfirmation('completed');
    expect(text).toContain('sessão e o feedback serão apagados');
    expect(text).toContain('ajustes');
  });

  it('é mais simples ao reabrir um treino não realizado', () => {
    expect(undoConfirmation('skipped')).toContain('Reabrir');
  });
});
