'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  Bike,
  CheckCircle2,
  LoaderCircle,
  MessageSquareHeart,
  Send,
} from 'lucide-react';
import { AccountActions } from '@/components/account-actions';
import { apiRequest } from '@/lib/api';

type User = { display_name: string };
type Category = 'experience' | 'bug' | 'suggestion';

const categories: Array<{ value: Category; label: string; description: string }> = [
  { value: 'experience', label: 'Minha experiência', description: 'Como foi usar o app e seguir o treino.' },
  { value: 'bug', label: 'Encontrei um problema', description: 'Algo não funcionou como deveria.' },
  { value: 'suggestion', label: 'Tenho uma sugestão', description: 'Uma ideia para deixar o Cadência melhor.' },
];

export default function FeedbackPage() {
  const [user, setUser] = useState<User | null>(null);
  const [category, setCategory] = useState<Category>('experience');
  const [rating, setRating] = useState(0);
  const [message, setMessage] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    apiRequest<{ user: User }>('/v1/me')
      .then(({ user: account }) => setUser(account))
      .catch(() => {
        window.location.href = '/entrar';
      })
      .finally(() => setLoading(false));
  }, []);

  async function submit(event: { preventDefault: () => void }) {
    event.preventDefault();
    setSending(true);
    setError('');
    setSubmitted(false);
    try {
      await apiRequest('/v1/feedback', {
        method: 'POST',
        body: JSON.stringify({ category, rating, message }),
      });
      setSubmitted(true);
      setRating(0);
      setMessage('');
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Não foi possível registrar seu feedback.');
    } finally {
      setSending(false);
    }
  }

  if (loading || !user) {
    return <main className="profile-loading"><LoaderCircle className="spin" />Carregando feedback…</main>;
  }

  return (
    <main className="feedback-shell">
      <header className="profile-topbar">
        <Link href="/" className="account-brand dark"><span><Bike size={19} /></span>cadência</Link>
        <AccountActions label="FEEDBACK" name={user.display_name} />
      </header>
      <section className="feedback-content">
        <Link href="/" className="back-link"><ArrowLeft size={15} />Voltar ao painel</Link>
        <header className="feedback-heading">
          <p>AJUDE A EVOLUIR O CADÊNCIA</p>
          <h1>Como está sendo sua experiência?</h1>
          <span>Seu relato ajuda a corrigir detalhes e construir treinos cada vez mais claros para quem pedala.</span>
        </header>
        <div className="feedback-layout">
          <section className="feedback-guide">
            <span className="feedback-icon"><MessageSquareHeart size={24} /></span>
            <h2>Conte o que você percebeu.</h2>
            <p>Você pode falar sobre a criação do perfil, o plano, a explicação dos treinos ou qualquer detalhe da sua experiência.</p>
            <ul>
              <li><CheckCircle2 size={17} /><span>Seja específico: isso facilita a investigação.</span></li>
              <li><CheckCircle2 size={17} /><span>Não inclua dados sensíveis ou informações de outras pessoas.</span></li>
              <li><CheckCircle2 size={17} /><span>O registro fica vinculado à sua conta para podermos organizar os relatos.</span></li>
            </ul>
          </section>
          <form className="feedback-form" onSubmit={submit}>
            <h2>Enviar feedback</h2>
            <p>Leva menos de um minuto. Obrigado por ajudar a testar o app.</p>
            <fieldset className="feedback-categories">
              <legend>Sobre o que você quer falar?</legend>
              <div>
                {categories.map((item) => (
                  <label key={item.value} htmlFor={`feedback-category-${item.value}`} aria-label={item.label} className={category === item.value ? 'selected' : ''}>
                    <input id={`feedback-category-${item.value}`} type="radio" name="category" value={item.value} checked={category === item.value} onChange={() => setCategory(item.value)} />
                    <span><strong>{item.label}</strong><small>{item.description}</small></span>
                  </label>
                ))}
              </div>
            </fieldset>
            <fieldset className="feedback-rating">
              <legend>Como você avalia a experiência? <small>1 = ruim · 5 = excelente</small></legend>
              <div aria-label="Nota da experiência">
                {[1, 2, 3, 4, 5].map((value) => (
                  <button key={value} type="button" className={rating === value ? 'selected' : ''} aria-pressed={rating === value} onClick={() => setRating(value)}>{value}</button>
                ))}
              </div>
            </fieldset>
            <div className="feedback-message">
              <label htmlFor="feedback-message">O que você gostaria de contar?</label>
              <textarea id="feedback-message" value={message} onChange={(event) => setMessage(event.target.value)} minLength={10} maxLength={2000} placeholder="O que funcionou bem? O que poderia ficar mais claro?" required />
              <small>{message.length}/2000 caracteres · mínimo de 10</small>
            </div>
            {error && <p className="form-error" role="alert">{error}</p>}
            {submitted && <p className="feedback-success" aria-live="polite"><CheckCircle2 size={17} />Feedback registrado. Obrigado por compartilhar!</p>}
            <button type="submit" disabled={sending || rating === 0 || message.trim().length < 10}>
              {sending ? <><LoaderCircle className="spin" size={17} />Enviando…</> : <><Send size={17} />Enviar feedback</>}
            </button>
          </form>
        </div>
      </section>
    </main>
  );
}
