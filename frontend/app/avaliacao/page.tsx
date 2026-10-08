'use client';

import { useEffect, useState } from 'react';
import { Activity, AlertTriangle, CalendarClock, CheckCircle2, Clock3, LoaderCircle, TrendingUp } from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { useLocale, useMessages } from '@/components/locale-provider';
import { ZoneHelp } from '@/components/zone-help';
import { defineMessages, formatDecimal, INTL_LOCALE } from '@/lib/i18n';
import { ZONES, zoneLabel, zoneName, zoneTalk } from '@/lib/zones';
import {
  describeDrift,
  efficiencyChange,
  fieldsFromRide,
  fitsAssessment,
  formatChange,
  formatEfficiency,
  halvesError,
  numbersPayload,
  reassessmentDate,
  reassessmentStatus,
  reassessRangeText,
  type Assessment,
  type ImportedRide,
} from '@/lib/assessment';

type User = { display_name: string };
type ImportedActivity = ImportedRide & { id: string; started_at: string };

const messages = defineMessages({
  pt: {
    resultPain: 'Você relatou dor. O app não usará este resultado para progredir intensidade; priorize recuperação e orientação profissional se a dor persistir.',
    resultEligible: 'Referência concluída sem sinal de alerta. Você está apto a receber treinos de qualidade quando o seu perfil e a sua meta pedirem, sempre com as regras de segurança.',
    resultNotEligible: 'Resultado salvo como referência. O motor continuará com progressão conservadora. Para ficar apto, faça o pedal de pelo menos 18 minutos, em Z2 ou Z3 e sem dor.',
    loadFailed: 'Não foi possível carregar sua avaliação.',
    saveFailed: 'Não foi possível registrar a avaliação.',
    loading: 'Carregando sua avaliação…',
    kicker: 'AVALIAÇÃO',
    title: 'Seu pedal de referência.',
    intro: 'Um pedal contínuo em Z2, com esforço controlado, para o app conhecer a sua base aeróbica e acompanhar a sua evolução. Não é exame médico nem teste máximo.',
    purpose: 'Para que serve',
    aptTitle: 'Apto a progredir:',
    aptText: 'sem dor, com pelo menos 18 minutos e esforço até Z3, você fica apto. Isso libera os treinos de qualidade (intervalados e limiar) para quem é de nível avançado e tem meta de desempenho ou prova. Para os outros perfis o plano não muda.',
    progressTitle: 'Sua evolução:',
    progressText: 'se você informar frequência cardíaca, potência ou distância, o app calcula a eficiência aeróbica e compara com a avaliação anterior.',
    redoTitle: 'Quando refazer:',
    redoText: (range: string) => `a cada ${range}, nas mesmas condições (mesmo percurso, ou rolo, e descansado).`,
    recorded: 'Avaliação registrada',
    recordedOn: (date: string) => ` em ${date}`,
    aptSuffix: ' · apto a progredir',
    efficiency: 'Eficiência aeróbica:',
    thanPrevious: 'que a anterior',
    cannotCompare: 'sem como comparar com a anterior (dados diferentes)',
    drift: 'Deriva da frequência cardíaca:',
    redoNow: 'Já dá para refazer a avaliação e ver como você evoluiu.',
    redoFrom: (date: string, days: number) => `Refaça a partir de ${date} (em ${days} dias).`,
    howTo: 'Como realizar',
    step1Title: 'Aqueça por 5 minutos',
    step1Text: 'Pedale leve e confortável.',
    step2Title: 'Pedale de forma contínua em Z2',
    step2Text: 'Esforço em que dá para conversar em frases completas, de 15 a 30 minutos (20 é o ideal). Evite descidas longas e paradas.',
    step3Title: 'Desaqueça',
    step3Text: 'Reduza o ritmo por 5 minutos antes de registrar como se sentiu.',
    tip: 'Para medir a deriva, anote (ou veja no aparelho) a frequência cardíaca média da primeira e da segunda metade do pedal.',
    stopNow: 'Interrompa imediatamente',
    stopText: ' se houver dor, tontura, falta de ar incomum, mal-estar ou outro sintoma preocupante.',
    evidence: 'Base científica:',
    evidenceText: ', sobre prescrição submáxima orientada por esforço percebido. A leitura da eficiência e da deriva é uma referência do Cadência, não um diagnóstico.',
    formTitle: 'Registrar resultado',
    formIntro: 'Faça o registro após o pedal. Não tente compensar ou alcançar um número específico.',
    fillFromRide: 'Preencher com um pedal importado (opcional)',
    chooseActivity: 'Escolher uma atividade…',
    duration: 'Duração contínua (minutos)',
    zoneQuestion: 'Em que zona você pedalou?',
    zoneHint: 'O pedido era Z2. Responda o que você sentiu, mesmo que tenha sido outra zona.',
    planned: 'pedido',
    numbers: 'Números do pedal',
    optional: 'opcionais',
    averageHr: 'FC média (bpm)',
    averagePower: 'Potência média (W)',
    distance: 'Distância (km)',
    firstHalf: 'FC média, 1ª metade',
    secondHalf: 'FC média, 2ª metade',
    example: (value: string) => `Ex.: ${value}`,
    numbersHint: 'Com FC e potência (ou FC e distância) o app calcula a eficiência; com as duas metades, a deriva. Nada disso muda o "apto a progredir".',
    pain: 'Relatei dor durante ou após o pedal',
    notes: 'Observações opcionais',
    confirm: 'Li as orientações e não realizei esforço máximo.',
    saving: 'Salvando…',
    save: 'Salvar avaliação',
    history: 'Suas avaliações',
    painReported: 'Dor relatada',
    apt: 'Apto',
    notApt: 'Não apto',
    driftShort: 'deriva',
  },
  en: {
    resultPain: 'You reported pain. The app will not use this result to progress intensity; prioritize recovery and see a professional if the pain persists.',
    resultEligible: 'Reference completed with no warning signs. You are cleared to get quality workouts when your profile and goal call for them, always within the safety rules.',
    resultNotEligible: 'Result saved as a reference. The engine will keep a conservative progression. To get cleared, ride for at least 18 minutes, in Z2 or Z3 and without pain.',
    loadFailed: 'Your assessment could not be loaded.',
    saveFailed: 'The assessment could not be saved.',
    loading: 'Loading your assessment…',
    kicker: 'ASSESSMENT',
    title: 'Your reference ride.',
    intro: 'A continuous Z2 ride, at a controlled effort, so the app can learn your aerobic base and follow your progress. It is not a medical exam or a maximal test.',
    purpose: 'What it is for',
    aptTitle: 'Cleared to progress:',
    aptText: 'with no pain, at least 18 minutes and effort up to Z3, you are cleared. This unlocks quality workouts (intervals and threshold) for advanced riders with a performance or race goal. For other profiles the plan does not change.',
    progressTitle: 'Your progress:',
    progressText: 'if you enter heart rate, power or distance, the app calculates your aerobic efficiency and compares it with the previous assessment.',
    redoTitle: 'When to redo it:',
    redoText: (range: string) => `every ${range}, under the same conditions (same route, or the trainer, and well rested).`,
    recorded: 'Assessment recorded',
    recordedOn: (date: string) => ` on ${date}`,
    aptSuffix: ' · cleared to progress',
    efficiency: 'Aerobic efficiency:',
    thanPrevious: 'than the previous one',
    cannotCompare: 'cannot be compared with the previous one (different data)',
    drift: 'Heart rate drift:',
    redoNow: 'You can now redo the assessment and see how you have progressed.',
    redoFrom: (date: string, days: number) => `Redo it from ${date} (in ${days} days).`,
    howTo: 'How to do it',
    step1Title: 'Warm up for 5 minutes',
    step1Text: 'Ride easy and comfortably.',
    step2Title: 'Ride continuously in Z2',
    step2Text: 'An effort where you can talk in full sentences, for 15 to 30 minutes (20 is ideal). Avoid long descents and stops.',
    step3Title: 'Cool down',
    step3Text: 'Ease off for 5 minutes before logging how you felt.',
    tip: 'To measure drift, write down (or check on your device) the average heart rate for the first and second halves of the ride.',
    stopNow: 'Stop immediately',
    stopText: ' if you have pain, dizziness, unusual shortness of breath, feel unwell or have any other worrying symptom.',
    evidence: 'Scientific basis:',
    evidenceText: ', on submaximal prescription guided by perceived effort. The efficiency and drift reading is a Cadência reference, not a diagnosis.',
    formTitle: 'Log the result',
    formIntro: "Log it after the ride. Don't try to make up for anything or hit a specific number.",
    fillFromRide: 'Fill in from an imported ride (optional)',
    chooseActivity: 'Choose an activity…',
    duration: 'Continuous duration (minutes)',
    zoneQuestion: 'Which zone did you ride in?',
    zoneHint: 'The target was Z2. Answer what you felt, even if it was another zone.',
    planned: 'target',
    numbers: 'Ride numbers',
    optional: 'optional',
    averageHr: 'Average HR (bpm)',
    averagePower: 'Average power (W)',
    distance: 'Distance (km)',
    firstHalf: 'Average HR, 1st half',
    secondHalf: 'Average HR, 2nd half',
    example: (value: string) => `E.g.: ${value}`,
    numbersHint: 'With HR and power (or HR and distance) the app calculates efficiency; with both halves, drift. None of this changes "cleared to progress".',
    pain: 'I had pain during or after the ride',
    notes: 'Optional notes',
    confirm: 'I read the guidance and did not make a maximal effort.',
    saving: 'Saving…',
    save: 'Save assessment',
    history: 'Your assessments',
    painReported: 'Pain reported',
    apt: 'Cleared',
    notApt: 'Not cleared',
    driftShort: 'drift',
  },
});

export default function AssessmentPage() {
  const locale = useLocale();
  const t = useMessages(messages);
  const [user, setUser] = useState<User | null>(null);
  const [assessments, setAssessments] = useState<Assessment[]>([]);
  const [rides, setRides] = useState<ImportedActivity[]>([]);
  const [duration, setDuration] = useState('20');
  const [actualZone, setActualZone] = useState(2);
  const [painReported, setPainReported] = useState(false);
  const [averageHeartRate, setAverageHeartRate] = useState('');
  const [averagePower, setAveragePower] = useState('');
  const [distanceKm, setDistanceKm] = useState('');
  const [firstHalf, setFirstHalf] = useState('');
  const [secondHalf, setSecondHalf] = useState('');
  const [notes, setNotes] = useState('');
  const [confirmedSafe, setConfirmedSafe] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [now, setNow] = useState<Date | null>(null);

  useEffect(() => {
    Promise.all([
      apiRequest<{ user: User }>('/v1/me'),
      apiRequest<{ assessments: Assessment[] }>('/v1/assessments'),
      apiRequest<{ activities: ImportedActivity[] }>('/v1/activities/imported').catch(() => ({ activities: [] })),
    ])
      .then(([account, history, imported]) => {
        setUser(account.user);
        setNow(new Date());
        setAssessments(history.assessments);
        setRides(imported.activities.filter(fitsAssessment).slice(0, 10));
      })
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

  function fillFromRide(id: string) {
    const ride = rides.find((item) => item.id === id);
    if (!ride) return;
    const fields = fieldsFromRide(ride);
    setDuration(fields.duration);
    setAverageHeartRate(fields.averageHeartRate);
    setAveragePower(fields.averagePower);
    setDistanceKm(fields.distanceKm);
  }

  async function submit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();
    const halves = halvesError(firstHalf, secondHalf, locale);
    if (halves) {
      setError(halves);
      return;
    }
    setSaving(true);
    setError('');
    try {
      const result = await apiRequest<{ assessment: Assessment }>('/v1/assessments/submaximal', {
        method: 'POST',
        body: JSON.stringify({
          duration_minutes: Number(duration),
          actual_zone: actualZone,
          pain_reported: painReported,
          notes,
          ...numbersPayload({ averageHeartRate, averagePower, distanceKm, firstHalf, secondHalf }),
        }),
      });
      setAssessments((current) => [result.assessment, ...current].slice(0, 10));
      setNow(new Date());
      setAverageHeartRate('');
      setAveragePower('');
      setDistanceKm('');
      setFirstHalf('');
      setSecondHalf('');
      setNotes('');
      setConfirmedSafe(false);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : t.saveFailed);
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        {t.loading}
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  const dayFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: '2-digit', month: '2-digit', year: 'numeric' });
  const shortDayFormatter = new Intl.DateTimeFormat(INTL_LOCALE[locale], { day: '2-digit', month: '2-digit' });
  const formatDay = (value?: string) => (value ? dayFormatter.format(new Date(value)) : '');
  const signedPercent = (value: number) => `${value > 0 ? '+' : ''}${formatDecimal(value, 1, locale)}%`;
  const resultMessage = (item: Assessment) =>
    item.pain_reported ? t.resultPain : item.eligible_for_progression ? t.resultEligible : t.resultNotEligible;

  const assessment = assessments[0];
  const previous = assessments[1];
  const change = assessment ? efficiencyChange(assessment.efficiency, previous?.efficiency) : null;
  const drift = assessment?.heart_rate_drift_percent;
  const reassess = assessment?.completed_at && now ? reassessmentStatus(assessment.completed_at, now) : null;

  return (
    <main className="assessment-shell">
      <AppHeader name={user.display_name} />
      <section className="assessment-content">
        <header className="assessment-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>{t.intro}</span>
        </header>

        <section className="assessment-purpose">
          <h2>
            <TrendingUp size={18} /> {t.purpose}
          </h2>
          <ul>
            <li>
              <strong>{t.aptTitle}</strong> {t.aptText}
            </li>
            <li>
              <strong>{t.progressTitle}</strong> {t.progressText}
            </li>
            <li>
              <strong>{t.redoTitle}</strong> {t.redoText(reassessRangeText(locale))}
            </li>
          </ul>
        </section>

        {assessment && (
          <section className={`assessment-result ${assessment.eligible_for_progression ? 'eligible' : ''}`}>
            <CheckCircle2 size={21} />
            <div>
              <strong>
                {t.recorded}
                {assessment.completed_at ? t.recordedOn(formatDay(assessment.completed_at)) : ''}
                {assessment.eligible_for_progression && !assessment.pain_reported ? t.aptSuffix : ''}
              </strong>
              <p>{resultMessage(assessment)}</p>
              {assessment.efficiency && (
                <p className="assessment-number">
                  <b>{t.efficiency}</b> {formatEfficiency(assessment.efficiency, locale)}
                  {change !== null && <> · {formatChange(change, locale)} {t.thanPrevious}</>}
                  {change === null && previous && <> · {t.cannotCompare}</>}
                </p>
              )}
              {drift !== undefined && (
                <p className={`assessment-number drift-${describeDrift(drift, locale).tone}`}>
                  <b>{t.drift}</b> {signedPercent(drift)}. {describeDrift(drift, locale).text}
                </p>
              )}
              {reassess && (
                <p className="assessment-next">
                  <CalendarClock size={15} />
                  {reassess.due
                    ? t.redoNow
                    : t.redoFrom(shortDayFormatter.format(reassessmentDate(assessment.completed_at ?? '')), reassess.daysLeft)}
                </p>
              )}
            </div>
          </section>
        )}

        <div className="assessment-layout">
          <section className="assessment-card">
            <span className="assessment-icon">
              <Activity size={22} />
            </span>
            <h2>{t.howTo}</h2>
            <ol>
              <li>
                <b>1</b>
                <span>
                  <strong>{t.step1Title}</strong>
                  {t.step1Text}
                </span>
              </li>
              <li>
                <b>2</b>
                <span>
                  <strong>{t.step2Title}</strong>
                  {t.step2Text}
                </span>
              </li>
              <li>
                <b>3</b>
                <span>
                  <strong>{t.step3Title}</strong>
                  {t.step3Text}
                </span>
              </li>
            </ol>
            <p className="assessment-tip">{t.tip}</p>
            <div className="assessment-warning">
              <AlertTriangle size={17} />
              <p>
                <strong>{t.stopNow}</strong>
                {t.stopText}
              </p>
            </div>
            <p className="assessment-evidence">
              {t.evidence}{' '}
              <a href="https://pubmed.ncbi.nlm.nih.gov/8668467/" target="_blank" rel="noreferrer">
                Dunbar, Kalinski {locale === 'en' ? 'and' : 'e'} Robertson (1996)
              </a>
              {t.evidenceText}
            </p>
          </section>

          <form className="assessment-form" onSubmit={submit}>
            <h2>{t.formTitle}</h2>
            <p>{t.formIntro}</p>

            {rides.length > 0 && (
              <label>
                <span>{t.fillFromRide}</span>
                <select defaultValue="" onChange={(event) => fillFromRide(event.target.value)}>
                  <option value="">{t.chooseActivity}</option>
                  {rides.map((ride) => (
                    <option value={ride.id} key={ride.id}>
                      {shortDayFormatter.format(new Date(ride.started_at))} · {Math.round(ride.moving_seconds / 60)} min
                      {ride.distance_km > 0 ? ` · ${formatDecimal(ride.distance_km, 1, locale)} km` : ''}
                    </option>
                  ))}
                </select>
              </label>
            )}

            <label>
              <span>{t.duration}</span>
              <input
                type="number"
                min="15"
                max="30"
                step="1"
                inputMode="numeric"
                value={duration}
                onChange={(event) => setDuration(event.target.value)}
                required
              />
            </label>

            <fieldset className="zone-choice">
              <legend>
                {t.zoneQuestion} <ZoneHelp compact />
              </legend>
              <small>{t.zoneHint}</small>
              <div className="zone-cards">
                {ZONES.map((zone) => (
                  <label key={zone.number} className={['zone-card', actualZone === zone.number ? 'selected' : ''].filter(Boolean).join(' ')}>
                    <input
                      type="radio"
                      name="assessment-zone"
                      checked={actualZone === zone.number}
                      onChange={() => setActualZone(zone.number)}
                    />
                    <strong>
                      Z{zone.number} · {zoneName(zone, locale)}
                      {zone.number === 2 && <em className="zone-planned">{t.planned}</em>}
                    </strong>
                    <span>{zoneTalk(zone, locale)}</span>
                  </label>
                ))}
              </div>
            </fieldset>

            <fieldset className="assessment-numbers">
              <legend>
                {t.numbers} <small>{t.optional}</small>
              </legend>
              <div className="assessment-numbers-grid">
                <label>
                  {t.averageHr}
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={averageHeartRate} onChange={(event) => setAverageHeartRate(event.target.value)} placeholder={t.example('140')} />
                </label>
                <label>
                  {t.averagePower}
                  <input type="number" min="0" max="2000" step="1" inputMode="numeric" value={averagePower} onChange={(event) => setAveragePower(event.target.value)} placeholder={t.example('165')} />
                </label>
                <label>
                  {t.distance}
                  <input type="number" min="0" max="500" step="0.01" inputMode="decimal" value={distanceKm} onChange={(event) => setDistanceKm(event.target.value)} placeholder={t.example(formatDecimal(9.5, 1, locale))} />
                </label>
                <span aria-hidden="true" />
                <label>
                  {t.firstHalf}
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={firstHalf} onChange={(event) => setFirstHalf(event.target.value)} placeholder={t.example('135')} />
                </label>
                <label>
                  {t.secondHalf}
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={secondHalf} onChange={(event) => setSecondHalf(event.target.value)} placeholder={t.example('144')} />
                </label>
              </div>
              <small>{t.numbersHint}</small>
            </fieldset>

            <label className="assessment-check">
              <input type="checkbox" checked={painReported} onChange={(event) => setPainReported(event.target.checked)} />
              <span>{t.pain}</span>
            </label>
            <label>
              <span>{t.notes}</span>
              <textarea maxLength={1000} value={notes} onChange={(event) => setNotes(event.target.value)} />
            </label>
            <label className="assessment-check">
              <input type="checkbox" checked={confirmedSafe} onChange={(event) => setConfirmedSafe(event.target.checked)} required />
              <span>{t.confirm}</span>
            </label>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <button type="submit" disabled={saving}>
              {saving ? <LoaderCircle className="spin" size={16} /> : <Clock3 size={16} />}
              {saving ? t.saving : t.save}
            </button>
          </form>
        </div>

        {assessments.length > 0 && (
          <section className="assessment-history">
            <h2>{t.history}</h2>
            <ul>
              {assessments.map((item, index) => {
                const older = assessments[index + 1];
                const itemChange = efficiencyChange(item.efficiency, older?.efficiency);
                return (
                  <li key={item.id}>
                    <strong>{formatDay(item.completed_at)}</strong>
                    <span>
                      {zoneLabel(item.actual_rpe, locale)} · {item.duration_minutes} min
                    </span>
                    <span className={item.eligible_for_progression && !item.pain_reported ? 'assessment-apt' : 'assessment-notapt'}>
                      {item.pain_reported ? t.painReported : item.eligible_for_progression ? t.apt : t.notApt}
                    </span>
                    {item.efficiency && (
                      <span>
                        {formatEfficiency(item.efficiency, locale)}
                        {itemChange !== null && ` (${formatChange(itemChange, locale)})`}
                      </span>
                    )}
                    {item.heart_rate_drift_percent !== undefined && (
                      <span>
                        {t.driftShort} {signedPercent(item.heart_rate_drift_percent)}
                      </span>
                    )}
                  </li>
                );
              })}
            </ul>
          </section>
        )}
      </section>
    </main>
  );
}
