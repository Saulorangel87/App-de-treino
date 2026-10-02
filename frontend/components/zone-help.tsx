import { CircleHelp } from 'lucide-react';
import {
  ZONES,
  ZONE_DISCLAIMER,
  formatHeartRateRange,
  formatPowerRange,
  zoneForRpe,
  zoneRanges,
  type ZoneReference,
} from '@/lib/zones';

/** Ajuda curta sobre as cinco zonas de esforço (substitui a ajuda do RPE). */
export function ZoneHelp({ compact = false }: { compact?: boolean }) {
  return (
    <span className="rpe-help-wrap">
      <button
        type="button"
        className={compact ? 'rpe-help compact' : 'rpe-help'}
        aria-label="Zonas de esforço: de 1, recuperação, a 5, esforço intenso. Passe o mouse ou foque para ver o que cada uma significa."
      >
        <CircleHelp size={compact ? 13 : 15} />
      </button>
      <span className="rpe-tooltip zone-tooltip" role="tooltip">
        <strong>Zonas de esforço</strong>
        {ZONES.map((zone) => (
          <span key={zone.number}>
            <b>
              Z{zone.number} {zone.name}
            </b>{' '}
            {zone.talk}
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
export function ZoneSummary({
  rpe,
  reference,
}: {
  rpe: number;
  reference: ZoneReference;
}) {
  const zone = zoneForRpe(rpe);
  const ranges = zoneRanges(zone, reference);
  const hasRanges = Boolean(ranges.heartRate || ranges.power);
  return (
    <div className="zone-summary" data-zone={zone.number}>
      <div className="zone-summary-head">
        <strong>
          Z{zone.number} · {zone.name}
        </strong>
        {ranges.heartRate && <span>{formatHeartRateRange(ranges.heartRate)}</span>}
        {ranges.power && <span>{formatPowerRange(ranges.power)}</span>}
      </div>
      <p>{zone.talk}</p>
      {!hasRanges && (
        <small>
          Sem sensor, guie-se pela conversa. Informe sua frequência cardíaca
          máxima (ou o FTP) no perfil para ver a faixa em batimentos (ou watts).
        </small>
      )}
      <small className="zone-disclaimer">{ZONE_DISCLAIMER}</small>
    </div>
  );
}
