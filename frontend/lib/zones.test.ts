import { describe, expect, it } from 'vitest';
import {
  ZONES,
  formatHeartRateRange,
  formatPowerRange,
  rpeForZone,
  zoneForRpe,
  zoneLabel,
  zoneRanges,
  zoneReferenceFrom,
} from './zones';

describe('zoneForRpe', () => {
  it('usa os limites de cada zona', () => {
    const cases: Array<[number, number]> = [
      [1, 1], [3.5, 1], [3.6, 2], [5, 2], [5.5, 3], [6.5, 3], [7, 4], [7.5, 4], [8, 5], [10, 5], [12, 5],
    ];
    for (const [rpe, zone] of cases) expect(zoneForRpe(rpe).number, `RPE ${rpe}`).toBe(zone);
  });

  it('cada RPE que o motor prescreve cai na zona prevista', () => {
    expect(zoneForRpe(3.5).number).toBe(1); // recuperação e giro leve protegido
    expect(zoneForRpe(4).number).toBe(2); // giro de base
    expect(zoneForRpe(5).number).toBe(2); // endurance e pedal longo
    expect(zoneForRpe(6).number).toBe(3); // tempo e intervalos moderados
    expect(zoneForRpe(6.5).number).toBe(3); // subidas controladas
    expect(zoneForRpe(7).number).toBe(4); // sweet spot
    expect(zoneForRpe(7.5).number).toBe(4); // intervalos de limiar
    expect(zoneForRpe(8).number).toBe(5); // intervalos intensos
  });
});

describe('rpeForZone', () => {
  it('devolve o RPE gravado e volta para a mesma zona', () => {
    for (const zone of ZONES) {
      const rpe = rpeForZone(zone.number);
      expect(rpe).toBeDefined();
      expect(zoneForRpe(rpe as number).number).toBe(zone.number);
    }
    expect(rpeForZone(0)).toBeUndefined();
    expect(rpeForZone(6)).toBeUndefined();
  });
});

describe('zoneLabel', () => {
  it('junta número e nome', () => {
    expect(zoneLabel(4.5, 'pt')).toBe('Z2 · Resistência');
    expect(zoneLabel(7, 'pt')).toBe('Z4 · Limiar');
    expect(zoneLabel(4.5, 'en')).toBe('Z2 · Endurance');
    expect(zoneLabel(7, 'en')).toBe('Z4 · Threshold');
  });
});

describe('zoneRanges', () => {
  it('sem referências, não há faixas (o app mostra o teste da conversa)', () => {
    expect(zoneRanges(ZONES[1], {})).toEqual({});
    expect(zoneRanges(ZONES[1], { maxHeartRate: 0, ftp: 0 })).toEqual({});
  });

  it('calcula as faixas de batimentos sem sobreposição entre zonas', () => {
    const max = 185;
    const ranges = ZONES.map((zone) => zoneRanges(zone, { maxHeartRate: max }).heartRate!);
    expect(ranges[0]).toEqual({ low: 93, high: 111 });
    expect(ranges[1]).toEqual({ low: 112, high: 130 });
    expect(ranges[4].high).toBe(185);
    for (let i = 1; i < ranges.length; i++) {
      expect(ranges[i].low).toBe(ranges[i - 1].high + 1);
    }
    expect(formatHeartRateRange(ranges[1])).toBe('112–130 bpm');
  });

  it('calcula as faixas de potência, com a primeira e a última abertas', () => {
    const ftp = 200;
    const [z1, z2, z3, z4, z5] = ZONES.map((zone) => zoneRanges(zone, { ftp }).power!);
    expect(z1).toEqual({ low: undefined, high: 110 });
    expect(z2).toEqual({ low: 111, high: 150 });
    expect(z3).toEqual({ low: 151, high: 180 });
    expect(z4).toEqual({ low: 181, high: 210 });
    expect(z5).toEqual({ low: 211, high: undefined });
    expect(formatPowerRange(z1, 'pt')).toBe('até 110 W');
    expect(formatPowerRange(z2, 'pt')).toBe('111–150 W');
    expect(formatPowerRange(z5, 'pt')).toBe('acima de 210 W');
    expect(formatPowerRange(z1, 'en')).toBe('up to 110 W');
    expect(formatPowerRange(z5, 'en')).toBe('above 210 W');
  });
});

describe('zoneReferenceFrom', () => {
  it('só usa a frequência e o FTP quando o atleta usa o sensor', () => {
    expect(zoneReferenceFrom(undefined)).toEqual({});
    expect(zoneReferenceFrom({ max_heart_rate: 190, uses_heart_rate: false, ftp: 220, uses_power: false })).toEqual({
      maxHeartRate: undefined,
      ftp: undefined,
    });
    expect(zoneReferenceFrom({ max_heart_rate: 190, uses_heart_rate: true, ftp: 220, uses_power: true })).toEqual({
      maxHeartRate: 190,
      ftp: 220,
    });
  });
});
