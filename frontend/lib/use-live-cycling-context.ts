'use client';

import { useEffect, useState } from 'react';
import { apiRequest } from './api';
import type { zoneReferenceFrom } from './zones';

export type LiveCyclingContext = NonNullable<Parameters<typeof zoneReferenceFrom>[0]>;

/**
 * Contexto de ciclismo atual do atleta. O plano guarda uma cópia do contexto de
 * quando foi gerado; as faixas das zonas usam esta leitura para refletir na hora
 * a frequência máxima, o limiar e o FTP que o atleta acabou de salvar.
 * Devolve undefined enquanto carrega ou se a leitura falhar (vale a cópia do plano).
 */
export function useLiveCyclingContext(): LiveCyclingContext | undefined {
  const [context, setContext] = useState<LiveCyclingContext>();
  useEffect(() => {
    let cancelled = false;
    apiRequest<{ onboarding: { cycling_context?: LiveCyclingContext } }>('/v1/onboarding')
      .then((result) => {
        if (!cancelled) setContext(result.onboarding.cycling_context);
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);
  return context;
}
