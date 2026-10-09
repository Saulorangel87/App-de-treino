'use client';

import { CircleHelp } from 'lucide-react';
import { defineMessages } from '@/lib/i18n';
import {
  ZONES,
  ZONE_DISCLAIMER,
  formatHeartRateRange,
  formatPowerRange,
  zoneForRpe,
  zoneName,
  zoneRanges,
  zoneTalk,
  type ZoneReference,
} from '@/lib/zones';
import { useLocale, useMessages } from './locale-provider';

const messages = defineMessages({
  pt: {
    helpLabel:
      'Zonas de esforço: de 1, recuperação, a 5, esforço intenso. Passe o mouse ou foque para ver o que cada uma significa.',
    title: 'Zonas de esforço',
    noSensor:
      'Sem sensor, guie-se pela conversa. Informe sua frequência cardíaca máxima (ou o FTP) no perfil para ver a faixa em batimentos (ou watts).',
  },
  en: {
    helpLabel:
      'Effort zones: from 1, recovery, to 5, high intensity. Hover or focus to see what each one means.',
    title: 'Effort zones',
    noSensor:
      'Without a sensor, use the talk test. Add your maximum heart rate (or FTP) to your profile to see the range in beats per minute (or watts).',
  },
});

/** Ajuda curta sobre as cinco zonas de esforço (substitui a ajuda do RPE). */
export function ZoneHelp({ compact = false }: { compact?: boolean }) {
  const locale = useLocale();
  const t = useMessages(messages);
  return (
    <span className="rpe-help-wrap">
      <button type="button" className={compact ? 'rpe-help compact' : 'rpe-help'} aria-label={t.helpLabel}>
        <CircleHelp size={compact ? 13 : 15} />
      </button>
      <span className="rpe-tooltip zone-tooltip" role="tooltip">
        <strong>{t.title}</strong>
        {ZONES.map((zone) => (
          <span key={zone.number}>
            <b>
              Z{zone.number} {zoneName(zone, locale)}
            </b>{' '}
            {zoneTalk(zone, locale)}
          </span>
        ))}
      </span>
    </span>
  );
}

/**
 * Resumo da zona de um treino: nome, faixa em batimentos e em watts (quando o
 * atleta informou a frequência máxima e/ou o FTP) e o teste da conversa, que serve
 * a quem pedala sem sensor.
 */
export function ZoneSummary({ rpe, reference }: { rpe: number; reference: ZoneReference }) {
  const locale = useLocale();
  const t = useMessages(messages);
  const zone = zoneForRpe(rpe);
  const ranges = zoneRanges(zone, reference);
  const hasRanges = Boolean(ranges.heartRate || ranges.power);
  return (
    <div className="zone-summary" data-zone={zone.number}>
      <div className="zone-summary-head">
        <strong>
          Z{zone.number} · {zoneName(zone, locale)}
        </strong>
        {ranges.heartRate && <span>{formatHeartRateRange(ranges.heartRate, locale)}</span>}
        {ranges.power && <span>{formatPowerRange(ranges.power, locale)}</span>}
      </div>
      <p>{zoneTalk(zone, locale)}</p>
      {!hasRanges && <small>{t.noSensor}</small>}
      <small className="zone-disclaimer">{ZONE_DISCLAIMER[locale]}</small>
    </div>
  );
}
