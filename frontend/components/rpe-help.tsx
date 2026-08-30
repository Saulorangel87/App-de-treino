import { CircleHelp } from 'lucide-react';

export function RpeHelp({ compact = false }: { compact?: boolean }) {
  return (
    <span className="rpe-help-wrap">
      <button
        type="button"
        className={compact ? 'rpe-help compact' : 'rpe-help'}
        aria-label="RPE significa percepção de esforço. A escala vai de 1, muito leve, até 10, esforço máximo."
      >
        <CircleHelp size={compact ? 13 : 15} />
      </button>
      <span className="rpe-tooltip" role="tooltip">
        <strong>RPE · Percepção de esforço</strong>
        <span>
          Escala de 1 a 10: 1 é muito leve, 5 é moderado, 7 é forte e 10
          representa esforço máximo.
        </span>
      </span>
    </span>
  );
}
