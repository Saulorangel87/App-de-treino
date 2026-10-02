import { describe, expect, it } from 'vitest';
import {
  describeDrift,
  efficiencyChange,
  fieldsFromRide,
  fitsAssessment,
  formatChange,
  formatEfficiency,
  halvesError,
  numbersPayload,
  reassessmentStatus,
  type AssessmentEfficiency,
} from './assessment';

const watts = (value: number): AssessmentEfficiency => ({ value, kind: 'power_per_bpm' });
const speed = (value: number): AssessmentEfficiency => ({ value, kind: 'speed_per_100bpm' });

describe('efficiencyChange', () => {
  it('compara com a avaliação anterior do mesmo tipo', () => {
    expect(efficiencyChange(watts(1.2), watts(1.0))).toBe(20);
    expect(efficiencyChange(watts(0.95), watts(1.0))).toBe(-5);
  });

  it('não compara tipos diferentes nem falta de dado', () => {
    expect(efficiencyChange(watts(1.2), speed(15))).toBeNull();
    expect(efficiencyChange(watts(1.2), undefined)).toBeNull();
    expect(efficiencyChange(undefined, watts(1))).toBeNull();
    expect(efficiencyChange(watts(1.2), watts(0))).toBeNull();
  });

  it('escreve a variação em português', () => {
    expect(formatChange(6.7)).toBe('6,7% melhor');
    expect(formatChange(-2)).toBe('2,0% pior');
    expect(formatChange(0)).toBe('igual');
  });

  it('escreve a eficiência com a unidade', () => {
    expect(formatEfficiency(watts(1.2))).toBe('1,20 W por bpm');
    expect(formatEfficiency(speed(20))).toBe('20,00 km/h a cada 100 bpm');
  });
});

describe('describeDrift', () => {
  it('separa estável, subiu um pouco e subiu bastante', () => {
    expect(describeDrift(0).tone).toBe('steady');
    expect(describeDrift(-2).tone).toBe('steady');
    expect(describeDrift(3).tone).toBe('steady');
    expect(describeDrift(4.2).tone).toBe('rising');
    expect(describeDrift(5).tone).toBe('rising');
    expect(describeDrift(5.1).tone).toBe('high');
  });
});

describe('reassessmentStatus', () => {
  const done = '2026-09-01T12:00:00Z';

  it('conta os dias que faltam para refazer', () => {
    expect(reassessmentStatus(done, new Date('2026-09-10T12:00:00Z'))).toEqual({ due: false, daysLeft: 19 });
  });

  it('marca como hora de refazer a partir de quatro semanas', () => {
    expect(reassessmentStatus(done, new Date('2026-09-29T12:00:00Z'))).toEqual({ due: true, daysLeft: 0 });
    expect(reassessmentStatus(done, new Date('2026-11-01T12:00:00Z'))).toEqual({ due: true, daysLeft: 0 });
  });
});

describe('pedal importado', () => {
  it('só serve se couber de 15 a 30 minutos', () => {
    expect(fitsAssessment({ moving_seconds: 14 * 60 })).toBe(false);
    expect(fitsAssessment({ moving_seconds: 15 * 60 })).toBe(true);
    expect(fitsAssessment({ moving_seconds: 30 * 60 + 20 })).toBe(true);
    expect(fitsAssessment({ moving_seconds: 45 * 60 })).toBe(false);
  });

  it('preenche o formulário e deixa em branco o que a atividade não tem', () => {
    expect(fieldsFromRide({ moving_seconds: 1230, distance_km: 9.456, average_heart_rate: 141.6, average_power_watts: 160 })).toEqual({
      duration: '21', averageHeartRate: '142', averagePower: '160', distanceKm: '9.46',
    });
    expect(fieldsFromRide({ moving_seconds: 1200, distance_km: 0 })).toEqual({
      duration: '20', averageHeartRate: '', averagePower: '', distanceKm: '',
    });
  });
});

describe('numbersPayload e halvesError', () => {
  it('manda só o que foi preenchido, aceitando vírgula', () => {
    expect(numbersPayload({ averageHeartRate: '140', averagePower: '', distanceKm: '9,5', firstHalf: '135', secondHalf: '144' })).toEqual({
      average_heart_rate: 140, average_power_w: undefined, distance_km: 9.5, heart_rate_first_half: 135, heart_rate_second_half: 144,
    });
    expect(Object.values(numbersPayload({ averageHeartRate: ' ', averagePower: 'abc', distanceKm: '', firstHalf: '', secondHalf: '' })).every((v) => v === undefined)).toBe(true);
  });

  it('exige as duas metades juntas', () => {
    expect(halvesError('', '')).toBe('');
    expect(halvesError('130', '140')).toBe('');
    expect(halvesError('130', '')).not.toBe('');
    expect(halvesError('', '140')).not.toBe('');
  });
});
