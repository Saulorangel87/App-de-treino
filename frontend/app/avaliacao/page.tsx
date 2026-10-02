'use client';

import { useEffect, useState } from 'react';
import { Activity, AlertTriangle, CalendarClock, CheckCircle2, Clock3, LoaderCircle, TrendingUp } from 'lucide-react';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { AppHeader } from '@/components/app-header';
import { ApiErrorState } from '@/components/api-error-state';
import { ZoneHelp } from '@/components/zone-help';
import { ZONES, zoneLabel } from '@/lib/zones';
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
  REASSESS_RANGE_TEXT,
  type Assessment,
  type ImportedRide,
} from '@/lib/assessment';

type User = { display_name: string };
type ImportedActivity = ImportedRide & { id: string; started_at: string };

const dayFormatter = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' });
const shortDayFormatter = new Intl.DateTimeFormat('pt-BR', { day: '2-digit', month: '2-digit' });

function formatDay(value?: string) {
  return value ? dayFormatter.format(new Date(value)) : '';
}

function resultMessage(assessment: Assessment): string {
  if (assessment.pain_reported) {
    return 'Você relatou dor. O app não usará este resultado para progredir intensidade; priorize recuperação e orientação profissional se a dor persistir.';
  }
  if (assessment.eligible_for_progression) {
    return 'Referência concluída sem sinal de alerta. Você está apto a receber treinos de qualidade quando o seu perfil e a sua meta pedirem, sempre com as regras de segurança.';
  }
  return 'Resultado salvo como referência. O motor continuará com progressão conservadora. Para ficar apto, faça o pedal de pelo menos 18 minutos, em Z2 ou Z3 e sem dor.';
}

export default function AssessmentPage() {
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
        setError(apiErrorMessage(caught, 'Não foi possível carregar sua avaliação.'));
      })
      .finally(() => setLoading(false));
  }, []);

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
    const halves = halvesError(firstHalf, secondHalf);
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
      setError(caught instanceof Error ? caught.message : 'Não foi possível registrar a avaliação.');
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        Carregando sua avaliação…
      </main>
    );
  }
  if (!user) return <ApiErrorState message={error || 'Não foi possível carregar sua avaliação.'} />;

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
          <p>AVALIAÇÃO</p>
          <h1>Seu pedal de referência.</h1>
          <span>
            Um pedal contínuo em Z2, com esforço controlado, para o app conhecer a sua base aeróbica e acompanhar a
            sua evolução. Não é exame médico nem teste máximo.
          </span>
        </header>

        <section className="assessment-purpose">
          <h2>
            <TrendingUp size={18} /> Para que serve
          </h2>
          <ul>
            <li>
              <strong>Apto a progredir:</strong> sem dor, com pelo menos 18 minutos e esforço até Z3, você fica
              apto. Isso libera os treinos de qualidade (intervalados e limiar) para quem é de nível avançado e tem
              meta de desempenho ou prova. Para os outros perfis o plano não muda.
            </li>
            <li>
              <strong>Sua evolução:</strong> se você informar frequência cardíaca, potência ou distância, o app
              calcula a eficiência aeróbica e compara com a avaliação anterior.
            </li>
            <li>
              <strong>Quando refazer:</strong> a cada {REASSESS_RANGE_TEXT}, nas mesmas condições (mesmo percurso, ou
              rolo, e descansado).
            </li>
          </ul>
        </section>

        {assessment && (
          <section className={`assessment-result ${assessment.eligible_for_progression ? 'eligible' : ''}`}>
            <CheckCircle2 size={21} />
            <div>
              <strong>
                Avaliação registrada{assessment.completed_at ? ` em ${formatDay(assessment.completed_at)}` : ''}
                {assessment.eligible_for_progression && !assessment.pain_reported ? ' · apto a progredir' : ''}
              </strong>
              <p>{resultMessage(assessment)}</p>
              {assessment.efficiency && (
                <p className="assessment-number">
                  <b>Eficiência aeróbica:</b> {formatEfficiency(assessment.efficiency)}
                  {change !== null && <> · {formatChange(change)} que a anterior</>}
                  {change === null && previous && <> · sem como comparar com a anterior (dados diferentes)</>}
                </p>
              )}
              {drift !== undefined && (
                <p className={`assessment-number drift-${describeDrift(drift).tone}`}>
                  <b>Deriva da frequência cardíaca:</b> {drift > 0 ? '+' : ''}
                  {drift.toFixed(1).replace('.', ',')}%. {describeDrift(drift).text}
                </p>
              )}
              {reassess && (
                <p className="assessment-next">
                  <CalendarClock size={15} />
                  {reassess.due
                    ? 'Já dá para refazer a avaliação e ver como você evoluiu.'
                    : `Refaça a partir de ${shortDayFormatter.format(reassessmentDate(assessment.completed_at ?? ''))} (em ${reassess.daysLeft} dias).`}
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
            <h2>Como realizar</h2>
            <ol>
              <li>
                <b>1</b>
                <span>
                  <strong>Aqueça por 5 minutos</strong>Pedale leve e confortável.
                </span>
              </li>
              <li>
                <b>2</b>
                <span>
                  <strong>Pedale de forma contínua em Z2</strong>
                  Esforço em que dá para conversar em frases completas, de 15 a 30 minutos (20 é o ideal). Evite
                  descidas longas e paradas.
                </span>
              </li>
              <li>
                <b>3</b>
                <span>
                  <strong>Desaqueça</strong>Reduza o ritmo por 5 minutos antes de registrar como se sentiu.
                </span>
              </li>
            </ol>
            <p className="assessment-tip">
              Para medir a deriva, anote (ou veja no aparelho) a frequência cardíaca média da primeira e da segunda
              metade do pedal.
            </p>
            <div className="assessment-warning">
              <AlertTriangle size={17} />
              <p>
                <strong>Interrompa imediatamente</strong> se houver dor, tontura, falta de ar incomum, mal-estar ou
                outro sintoma preocupante.
              </p>
            </div>
            <p className="assessment-evidence">
              Base científica:{' '}
              <a href="https://pubmed.ncbi.nlm.nih.gov/8668467/" target="_blank" rel="noreferrer">
                Dunbar, Kalinski e Robertson (1996)
              </a>
              , sobre prescrição submáxima orientada por esforço percebido. A leitura da eficiência e da deriva é
              uma referência do Cadência, não um diagnóstico.
            </p>
          </section>

          <form className="assessment-form" onSubmit={submit}>
            <h2>Registrar resultado</h2>
            <p>Faça o registro após o pedal. Não tente compensar ou alcançar um número específico.</p>

            {rides.length > 0 && (
              <label>
                <span>Preencher com um pedal importado (opcional)</span>
                <select defaultValue="" onChange={(event) => fillFromRide(event.target.value)}>
                  <option value="">Escolher uma atividade…</option>
                  {rides.map((ride) => (
                    <option value={ride.id} key={ride.id}>
                      {shortDayFormatter.format(new Date(ride.started_at))} · {Math.round(ride.moving_seconds / 60)} min
                      {ride.distance_km > 0 ? ` · ${ride.distance_km.toFixed(1).replace('.', ',')} km` : ''}
                    </option>
                  ))}
                </select>
              </label>
            )}

            <label>
              <span>Duração contínua (minutos)</span>
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
                Em que zona você pedalou? <ZoneHelp compact />
              </legend>
              <small>O pedido era Z2. Responda o que você sentiu, mesmo que tenha sido outra zona.</small>
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
                      Z{zone.number} · {zone.name}
                      {zone.number === 2 && <em className="zone-planned">pedido</em>}
                    </strong>
                    <span>{zone.talk}</span>
                  </label>
                ))}
              </div>
            </fieldset>

            <fieldset className="assessment-numbers">
              <legend>
                Números do pedal <small>opcionais</small>
              </legend>
              <div className="assessment-numbers-grid">
                <label>
                  FC média (bpm)
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={averageHeartRate} onChange={(event) => setAverageHeartRate(event.target.value)} placeholder="Ex.: 140" />
                </label>
                <label>
                  Potência média (W)
                  <input type="number" min="0" max="2000" step="1" inputMode="numeric" value={averagePower} onChange={(event) => setAveragePower(event.target.value)} placeholder="Ex.: 165" />
                </label>
                <label>
                  Distância (km)
                  <input type="number" min="0" max="500" step="0.01" inputMode="decimal" value={distanceKm} onChange={(event) => setDistanceKm(event.target.value)} placeholder="Ex.: 9,5" />
                </label>
                <span aria-hidden="true" />
                <label>
                  FC média, 1ª metade
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={firstHalf} onChange={(event) => setFirstHalf(event.target.value)} placeholder="Ex.: 135" />
                </label>
                <label>
                  FC média, 2ª metade
                  <input type="number" min="30" max="250" step="1" inputMode="numeric" value={secondHalf} onChange={(event) => setSecondHalf(event.target.value)} placeholder="Ex.: 144" />
                </label>
              </div>
              <small>
                Com FC e potência (ou FC e distância) o app calcula a eficiência; com as duas metades, a deriva. Nada
                disso muda o &quot;apto a progredir&quot;.
              </small>
            </fieldset>

            <label className="assessment-check">
              <input type="checkbox" checked={painReported} onChange={(event) => setPainReported(event.target.checked)} />
              <span>Relatei dor durante ou após o pedal</span>
            </label>
            <label>
              <span>Observações opcionais</span>
              <textarea maxLength={1000} value={notes} onChange={(event) => setNotes(event.target.value)} />
            </label>
            <label className="assessment-check">
              <input type="checkbox" checked={confirmedSafe} onChange={(event) => setConfirmedSafe(event.target.checked)} required />
              <span>Li as orientações e não realizei esforço máximo.</span>
            </label>
            {error && (
              <p className="form-error" role="alert">
                {error}
              </p>
            )}
            <button type="submit" disabled={saving}>
              {saving ? <LoaderCircle className="spin" size={16} /> : <Clock3 size={16} />}
              {saving ? 'Salvando…' : 'Salvar avaliação'}
            </button>
          </form>
        </div>

        {assessments.length > 0 && (
          <section className="assessment-history">
            <h2>Suas avaliações</h2>
            <ul>
              {assessments.map((item, index) => {
                const older = assessments[index + 1];
                const itemChange = efficiencyChange(item.efficiency, older?.efficiency);
                return (
                  <li key={item.id}>
                    <strong>{formatDay(item.completed_at)}</strong>
                    <span>
                      {zoneLabel(item.actual_rpe)} · {item.duration_minutes} min
                    </span>
                    <span className={item.eligible_for_progression && !item.pain_reported ? 'assessment-apt' : 'assessment-notapt'}>
                      {item.pain_reported ? 'Dor relatada' : item.eligible_for_progression ? 'Apto' : 'Não apto'}
                    </span>
                    {item.efficiency && (
                      <span>
                        {formatEfficiency(item.efficiency)}
                        {itemChange !== null && ` (${formatChange(itemChange)})`}
                      </span>
                    )}
                    {item.heart_rate_drift_percent !== undefined && (
                      <span>
                        deriva {item.heart_rate_drift_percent > 0 ? '+' : ''}
                        {item.heart_rate_drift_percent.toFixed(1).replace('.', ',')}%
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
