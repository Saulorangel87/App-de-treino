'use client';

import { useState } from 'react';
import { Activity, ArrowRight, Bike, CalendarDays, Check, ChevronRight, CircleHelp, Clock3, Gauge, Home, LineChart, Play, Settings, Sparkles, Zap } from 'lucide-react';

const week = [
  { day: 'SEG', date: '24', kind: 'Endurance', tone: 'mint', done: true },
  { day: 'TER', date: '25', kind: 'Descanso', tone: 'rest' },
  { day: 'QUA', date: '26', kind: 'Sweet spot', tone: 'lime', active: true },
  { day: 'QUI', date: '27', kind: 'Mobilidade', tone: 'blue' },
  { day: 'SEX', date: '28', kind: 'Descanso', tone: 'rest' },
  { day: 'SÁB', date: '29', kind: 'Longo', tone: 'orange' },
  { day: 'DOM', date: '30', kind: 'Recuperação', tone: 'mint' },
];

export default function HomePage() {
  const [feedbackOpen, setFeedbackOpen] = useState(false);
  const [checkinOpen, setCheckinOpen] = useState(false);
  const [readiness, setReadiness] = useState(82);
  const [message, setMessage] = useState('');
  function startWorkout() {
    setMessage('Treino iniciado — o cronômetro está valendo. Bom pedal!');
    window.setTimeout(() => setMessage(''), 4200);
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div className="brand" aria-label="Cadência"><span className="brand-mark"><Bike size={21} strokeWidth={2.4} /></span><span>cadência</span></div>
        <nav className="main-nav" aria-label="Navegação principal">
          <button className="nav-item active"><Home size={18} />Visão geral</button>
          <button className="nav-item"><CalendarDays size={18} />Meu plano</button>
          <button className="nav-item"><Activity size={18} />Atividades</button>
          <button className="nav-item"><LineChart size={18} />Evolução</button>
        </nav>
        <div className="sidebar-bottom">
          <div className="coach-note"><span className="coach-icon"><Sparkles size={15} /></span><p><strong>Plano adaptativo</strong>Seu feedback ajusta a próxima semana.</p></div>
          <button className="nav-item"><Settings size={18} />Configurações</button>
          <button className="profile-mini"><span className="avatar">SR</span><span><strong>Saulo Ribeiro</strong><small>Ciclismo · Intermediário</small></span><ChevronRight size={15} /></button>
        </div>
      </aside>

      <section className="workspace">
        <header className="topbar"><div><p className="eyebrow">QUARTA, 26 DE AGOSTO</p><h1>Bom dia, Saulo.</h1></div><div className="top-actions"><button className="help-button" aria-label="Ajuda"><CircleHelp size={19} /></button><button className="recovery-pill" onClick={() => setCheckinOpen(true)}><span className="status-dot" />Recuperação boa <strong>{readiness}%</strong></button></div></header>
        <div className="dashboard-grid">
          <section className="today-card">
            <div className="today-copy"><div className="section-label"><span>HOJE</span><i>•</i> BIKE INDOOR</div><h2>Sweet spot<br />progressivo</h2><p>Blocos sustentados para construir potência sem comprometer sua recuperação.</p><div className="metric-row"><span><Clock3 size={17} /><b>60</b> min</span><span><Gauge size={17} /><b>RPE 7</b> / 10</span><span><Zap size={17} /><b>72</b> pts</span></div><button className="start-button" onClick={startWorkout}><Play size={16} fill="currentColor" />Iniciar treino</button><button className="details-button" onClick={() => setFeedbackOpen(true)}>Ver estrutura <ArrowRight size={15} /></button></div>
            <div className="effort-visual" aria-label="Perfil de esforço do treino"><div className="effort-label"><span>PERFIL DE ESFORÇO</span><strong>3 blocos</strong></div><div className="bars">{[28,36,44,54,54,54,30,64,64,64,30,74,74,74,38,26].map((height,i) => <span key={i} style={{ height }} className={i > 3 && i < 15 && ![6,10].includes(i) ? 'work' : ''} />)}</div><div className="zone-scale"><span>Z1</span><span>Z2</span><span>Z3</span><span>Z4</span></div><div className="coach-tip"><Sparkles size={16} /><p><strong>Ajuste inteligente</strong>Volume reduzido em 8% devido ao sono abaixo da sua média.</p></div></div>
          </section>
          <aside className="readiness-card"><div className="card-heading"><div><span>PRONTIDÃO</span><h3>Você está bem.</h3></div><button aria-label="Mais informações">•••</button></div><div className="score-wrap"><svg viewBox="0 0 120 120" aria-hidden="true"><circle cx="60" cy="60" r="48" className="score-bg" /><circle cx="60" cy="60" r="48" className="score-ring" /></svg><div><strong>{readiness}</strong><span>/100</span></div></div><div className="readiness-list"><div><span>Sono</span><b>6h42</b><i className="warn">Abaixo</i></div><div><span>Fadiga</span><b>Baixa</b><i>Boa</i></div><div><span>Estresse</span><b>Moderado</b><i>Normal</i></div></div><button className="checkin-button" onClick={() => setCheckinOpen(true)}>Atualizar check-in <ArrowRight size={15} /></button></aside>
          <section className="week-card"><div className="week-header"><div><span className="section-kicker">SUA SEMANA</span><h3>Consistência vence intensidade isolada.</h3></div><div className="week-progress"><strong>3 de 5</strong><span>sessões concluídas</span></div></div><div className="week-days">{week.map((item) => <article key={item.day} className={`day-card ${item.active ? 'selected' : ''}`}><div><span>{item.day}</span><strong>{item.date}</strong></div><i className={`workout-dot ${item.tone}`}>{item.done ? <Check size={13} /> : item.active ? <Bike size={14} /> : ''}</i><p>{item.kind}</p>{item.active && <small>60 min</small>}</article>)}</div><div className="load-summary"><span>CARGA SEMANAL</span><div className="load-track"><i /></div><strong>218 <small>/ 340 pts</small></strong><em>Dentro da meta</em></div></section>
          <section className="insight-card"><div className="insight-icon"><Sparkles size={18} /></div><div><span>POR QUE ESTE TREINO?</span><h3>O estímulo certo, no momento certo.</h3><p>Seu histórico mostra boa resposta a blocos sustentados. Como a fadiga está baixa e você tem 60 minutos disponíveis, o sweet spot entrega o melhor equilíbrio entre estímulo e recuperação hoje.</p><button>Entender a decisão <ArrowRight size={15} /></button></div><div className="evidence-tag"><span>BASEADO EM</span><strong>3 regras do seu perfil</strong><small>+ 2 referências científicas</small></div></section>
        </div>
      </section>
      {feedbackOpen && <div className="modal-backdrop" role="presentation" onMouseDown={() => setFeedbackOpen(false)}><section className="workout-modal" role="dialog" aria-modal="true" aria-labelledby="workout-title" onMouseDown={(e) => e.stopPropagation()}><button className="modal-close" onClick={() => setFeedbackOpen(false)} aria-label="Fechar">×</button><span className="modal-kicker">TREINO DE HOJE · 60 MIN</span><h2 id="workout-title">Sweet spot progressivo</h2><p>Use a percepção de esforço como guia principal. Você deve conseguir sustentar o bloco com foco, sem chegar à exaustão.</p><ol className="workout-steps"><li><b>01</b><span><strong>Aquecimento</strong><small>12 min · RPE 2–4</small></span><em>Cadência livre</em></li><li><b>02</b><span><strong>Bloco principal</strong><small>3 × 10 min · RPE 7</small></span><em>3 min leves entre blocos</em></li><li><b>03</b><span><strong>Desaquecimento</strong><small>9 min · RPE 2</small></span><em>Giro confortável</em></li></ol><button className="modal-start" onClick={() => { setFeedbackOpen(false); startWorkout(); }}><Play size={16} fill="currentColor" />Começar agora</button></section></div>}
      {checkinOpen && <div className="modal-backdrop" role="presentation" onMouseDown={() => setCheckinOpen(false)}><section className="workout-modal checkin-modal" role="dialog" aria-modal="true" aria-labelledby="checkin-title" onMouseDown={(e) => e.stopPropagation()}><button className="modal-close" onClick={() => setCheckinOpen(false)} aria-label="Fechar">×</button><span className="modal-kicker">CHECK-IN DIÁRIO</span><h2 id="checkin-title">Como você acordou?</h2><p>Leva menos de um minuto. Suas respostas ajudam o plano a respeitar sua recuperação.</p><div className="checkin-question"><strong>Qualidade do sono</strong><div><button>Ruim</button><button className="chosen">Razoável</button><button>Boa</button></div></div><div className="checkin-question"><strong>Fadiga percebida</strong><div><button>Alta</button><button>Moderada</button><button className="chosen">Baixa</button></div></div><div className="checkin-question"><strong>Nível de estresse</strong><div><button>Alto</button><button className="chosen">Moderado</button><button>Baixo</button></div></div><button className="modal-start" onClick={() => { setReadiness(84); setCheckinOpen(false); setMessage('Check-in salvo. Seu treino de hoje continua adequado.'); window.setTimeout(() => setMessage(''), 4200); }}><Check size={16} />Salvar check-in</button><small className="safety-note">Se houver dor ou lesão, interrompa o treino e procure um profissional habilitado.</small></section></div>}
      {message && <div className="toast" role="status"><span><Check size={15} /></span>{message}</div>}
    </main>
  );
}
