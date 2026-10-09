'use client';

import { useEffect, useState } from 'react';
import { HeartPulse } from 'lucide-react';
import Link from 'next/link';
import { apiErrorMessage, apiRequest } from '@/lib/api';
import { defineMessages, INTL_LOCALE } from '@/lib/i18n';
import { useLocale, useMessages } from './locale-provider';

type Context = {
  uses_heart_rate?: boolean;
  uses_power?: boolean;
  max_heart_rate?: number;
  lthr?: number;
  ftp?: number;
  ftp_test_date?: string;
  ftp_protocol?: string;
  [key: string]: unknown;
};

type Suggestion = { value: number; activities: number; observed_on: string };
type Suggestions = { max_heart_rate?: Suggestion; ftp?: Suggestion };

const messages = defineMessages({
  pt: {
    title: 'Meus números',
    intro: 'Frequência máxima, limiar de frequência e FTP em um só lugar. Com eles, as zonas dos treinos aparecem em batimentos e watts.',
    maxHr: 'Frequência cardíaca máxima (bpm)',
    lthr: 'Limiar de frequência cardíaca (bpm)',
    ftp: 'FTP (watts)',
    ftpOn: (date: string) => `Informado em ${date}`,
    noSensor: 'Marque o uso de frequência cardíaca ou de potência no perfil para informar estes números.',
    editProfile: 'Abrir perfil',
    save: 'Salvar números',
    saving: 'Salvando…',
    saved: 'Números salvos. As faixas das zonas já usam os novos valores.',
    saveFailed: 'Não foi possível salvar seus números.',
    suggestionTitle: 'Sugestões das suas atividades',
    suggestMaxHr: (value: number, count: number, date: string) =>
      `Seu maior batimento foi ${value} bpm, em ${date} (${count} atividades com frequência).`,
    suggestFtp: (value: number, count: number, date: string) =>
      `Seu melhor esforço de 20 minutos, em ${date}, sugere um FTP de ${value} W (95% da média; ${count} atividades com potência).`,
    use: 'Usar este valor',
    suggestionNote:
      'As sugestões vêm só dos arquivos que você importou e nada é salvo sozinho: o valor entra nos campos acima e você confirma em "Salvar números". Um esforço de 20 minutos feito em treino leve pode subestimar o seu FTP.',
    hint: 'Valores plausíveis: frequência de 100 a 230 bpm (o limiar não passa da máxima) e FTP de 50 a 600 W.',
  },
  en: {
    title: 'My numbers',
    intro: 'Maximum heart rate, lactate threshold heart rate and FTP in one place. With them, workout zones show in beats per minute and watts.',
    maxHr: 'Maximum heart rate (bpm)',
    lthr: 'Lactate threshold heart rate (bpm)',
    ftp: 'FTP (watts)',
    ftpOn: (date: string) => `Entered on ${date}`,
    noSensor: 'Turn on heart rate or power in your profile to enter these numbers.',
    editProfile: 'Open profile',
    save: 'Save numbers',
    saving: 'Saving…',
    saved: 'Numbers saved. Zone ranges already use the new values.',
    saveFailed: 'Your numbers could not be saved.',
    suggestionTitle: 'Suggestions from your activities',
    suggestMaxHr: (value: number, count: number, date: string) =>
      `Your highest heart rate was ${value} bpm, on ${date} (${count} activities with heart rate).`,
    suggestFtp: (value: number, count: number, date: string) =>
      `Your best 20-minute effort, on ${date}, suggests an FTP of ${value} W (95% of the average; ${count} activities with power).`,
    use: 'Use this value',
    suggestionNote:
      'Suggestions come only from the files you imported and nothing is saved on its own: the value goes into the fields above and you confirm with "Save numbers". A 20-minute effort done in an easy workout may underestimate your FTP.',
    hint: 'Plausible values: heart rate from 100 to 230 bpm (the threshold cannot exceed the maximum) and FTP from 50 to 600 W.',
  },
});

function numberOrUndefined(value: string): number | undefined {
  const parsed = Number(value);
  return parsed > 0 ? Math.round(parsed) : undefined;
}

/** Cartão de Configurações com os números de referência do atleta e as sugestões tiradas das atividades importadas. */
export function ReferenceNumbersCard() {
  const locale = useLocale();
  const t = useMessages(messages);
  const [context, setContext] = useState<Context | null>(null);
  const [suggestions, setSuggestions] = useState<Suggestions>({});
  const [maxHr, setMaxHr] = useState('');
  const [lthr, setLthr] = useState('');
  const [ftp, setFtp] = useState('');
  const [ftpDate, setFtpDate] = useState<string | undefined>();
  const [ftpProtocol, setFtpProtocol] = useState<string | undefined>();
  const [status, setStatus] = useState<{ kind: 'ok' | 'error'; text: string } | null>(null);
  const [busy, setBusy] = useState(false);

  function fill(next: Context) {
    setContext(next);
    setMaxHr(next.max_heart_rate ? String(next.max_heart_rate) : '');
    setLthr(next.lthr ? String(next.lthr) : '');
    setFtp(next.ftp ? String(next.ftp) : '');
    setFtpDate(next.ftp_test_date);
    setFtpProtocol(next.ftp_protocol);
  }

  useEffect(() => {
    let cancelled = false;
    Promise.all([
      apiRequest<{ onboarding: { cycling_context?: Context } }>('/v1/onboarding'),
      apiRequest<{ suggestions: Suggestions }>('/v1/activities/reference-suggestions').catch(() => ({ suggestions: {} as Suggestions })),
    ])
      .then(([onboarding, reference]) => {
        if (cancelled) return;
        fill(onboarding.onboarding.cycling_context ?? {});
        setSuggestions(reference.suggestions ?? {});
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);

  // Sem perfil ou sem resposta da API o cartão simplesmente não aparece.
  if (!context) return null;

  const usesHr = Boolean(context.uses_heart_rate);
  const usesPower = Boolean(context.uses_power);
  const dayFormat = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: 'numeric', month: 'short', year: 'numeric' });
  const formatDay = (value: string) => dayFormat.format(new Date(`${value}T12:00:00`));

  const currentMax = numberOrUndefined(maxHr);
  const hrSuggestion =
    usesHr && suggestions.max_heart_rate && (currentMax === undefined || suggestions.max_heart_rate.value > currentMax)
      ? suggestions.max_heart_rate
      : undefined;
  const currentFtp = numberOrUndefined(ftp);
  const ftpSuggestion =
    usesPower && suggestions.ftp && (currentFtp === undefined || Math.abs(suggestions.ftp.value - currentFtp) >= 5)
      ? suggestions.ftp
      : undefined;

  async function save() {
    if (!context) return;
    setBusy(true);
    setStatus(null);
    try {
      const nextFtp = numberOrUndefined(ftp);
      const result = await apiRequest<{ cycling_context: Context }>('/v1/onboarding/cycling-context', {
        method: 'PUT',
        body: JSON.stringify({
          cycling_context: {
            ...context,
            max_heart_rate: usesHr ? numberOrUndefined(maxHr) : undefined,
            lthr: usesHr ? numberOrUndefined(lthr) : undefined,
            ftp: usesPower ? nextFtp : undefined,
            ftp_test_date: usesPower && nextFtp ? ftpDate : undefined,
            ftp_protocol: usesPower && nextFtp ? ftpProtocol : undefined,
          },
        }),
      });
      fill(result.cycling_context);
      setStatus({ kind: 'ok', text: t.saved });
    } catch (caught) {
      setStatus({ kind: 'error', text: apiErrorMessage(caught, t.saveFailed) });
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="settings-card settings-numbers-card">
      <span className="settings-icon">
        <HeartPulse size={23} />
      </span>
      <h2>{t.title}</h2>
      <p className="settings-card-intro">{t.intro}</p>
      {!usesHr && !usesPower ? (
        <p className="settings-numbers-empty">
          {t.noSensor} <Link href="/perfil">{t.editProfile}</Link>
        </p>
      ) : (
        <form
          className="settings-form"
          onSubmit={(event) => {
            event.preventDefault();
            void save();
          }}
        >
          {usesHr && (
            <>
              <label htmlFor="reference-max-hr">{t.maxHr}</label>
              <input id="reference-max-hr" type="number" inputMode="numeric" min={100} max={230} value={maxHr} onChange={(event) => setMaxHr(event.target.value)} />
              <label htmlFor="reference-lthr">{t.lthr}</label>
              <input id="reference-lthr" type="number" inputMode="numeric" min={100} max={230} value={lthr} onChange={(event) => setLthr(event.target.value)} />
            </>
          )}
          {usesPower && (
            <>
              <label htmlFor="reference-ftp">{t.ftp}</label>
              <input id="reference-ftp" type="number" inputMode="numeric" min={50} max={600} value={ftp} onChange={(event) => setFtp(event.target.value)} />
              {ftpDate && currentFtp !== undefined && <small>{t.ftpOn(formatDay(ftpDate))}</small>}
            </>
          )}
          <small className="settings-numbers-hint">{t.hint}</small>
          {(hrSuggestion || ftpSuggestion) && (
            <div className="settings-numbers-suggestions" aria-label={t.suggestionTitle}>
              <strong>{t.suggestionTitle}</strong>
              {hrSuggestion && (
                <p>
                  {t.suggestMaxHr(hrSuggestion.value, hrSuggestion.activities, formatDay(hrSuggestion.observed_on))}
                  <button type="button" className="settings-secondary-button" onClick={() => setMaxHr(String(hrSuggestion.value))}>
                    {t.use}
                  </button>
                </p>
              )}
              {ftpSuggestion && (
                <p>
                  {t.suggestFtp(ftpSuggestion.value, ftpSuggestion.activities, formatDay(ftpSuggestion.observed_on))}
                  <button
                    type="button"
                    className="settings-secondary-button"
                    onClick={() => {
                      setFtp(String(ftpSuggestion.value));
                      setFtpDate(ftpSuggestion.observed_on);
                      setFtpProtocol('20_minute');
                    }}
                  >
                    {t.use}
                  </button>
                </p>
              )}
              <small>{t.suggestionNote}</small>
            </div>
          )}
          {status && (
            <p className={status.kind === 'ok' ? 'form-notice' : 'form-error'} role="status">
              {status.text}
            </p>
          )}
          <button type="submit" className="settings-primary-button" disabled={busy}>
            {busy ? t.saving : t.save}
          </button>
        </form>
      )}
    </section>
  );
}
