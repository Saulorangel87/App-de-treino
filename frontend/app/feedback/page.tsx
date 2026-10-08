'use client';

import { useEffect, useState } from 'react';
import {
  CheckCircle2,
  LoaderCircle,
  MessageSquareHeart,
  Send,
} from 'lucide-react';
import { AppHeader } from '@/components/app-header';
import { ApiError, apiErrorMessage, apiRequest } from '@/lib/api';
import { ApiErrorState } from '@/components/api-error-state';
import { useMessages } from '@/components/locale-provider';
import { defineMessages } from '@/lib/i18n';

type User = { display_name: string };
type Category = 'experience' | 'bug' | 'suggestion';

const categories: Category[] = ['experience', 'bug', 'suggestion'];

const messages = defineMessages({
  pt: {
    categories: {
      experience: { label: 'Minha experiência', description: 'Como foi usar o app e seguir o treino.' },
      bug: { label: 'Encontrei um problema', description: 'Algo não funcionou como deveria.' },
      suggestion: { label: 'Tenho uma sugestão', description: 'Uma ideia para deixar o Cadência melhor.' },
    } as Record<Category, { label: string; description: string }>,
    loadFailed: 'Não foi possível carregar o feedback.',
    sendFailed: 'Não foi possível registrar seu feedback.',
    loading: 'Carregando feedback…',
    kicker: 'AJUDE A EVOLUIR O CADÊNCIA',
    title: 'Como está sendo sua experiência?',
    intro: 'Seu relato ajuda a corrigir detalhes e construir treinos cada vez mais claros para quem pedala.',
    guideTitle: 'Conte o que você percebeu.',
    guideText: 'Você pode falar sobre a criação do perfil, o plano, a explicação dos treinos ou qualquer detalhe da sua experiência.',
    tips: [
      'Seja específico: isso facilita a investigação.',
      'Não inclua dados sensíveis ou informações de outras pessoas.',
      'O registro fica vinculado à sua conta para podermos organizar os relatos.',
    ],
    formTitle: 'Enviar feedback',
    formIntro: 'Leva menos de um minuto. Obrigado por ajudar a testar o app.',
    topic: 'Sobre o que você quer falar?',
    rating: 'Como você avalia a experiência?',
    ratingScale: '1 = ruim · 5 = excelente',
    ratingLabel: 'Nota da experiência',
    message: 'O que você gostaria de contar?',
    messagePlaceholder: 'O que funcionou bem? O que poderia ficar mais claro?',
    count: (length: number) => `${length}/2000 caracteres · mínimo de 10`,
    sent: 'Feedback registrado. Obrigado por compartilhar!',
    sending: 'Enviando…',
    send: 'Enviar feedback',
  },
  en: {
    categories: {
      experience: { label: 'My experience', description: 'How it was to use the app and follow the workout.' },
      bug: { label: 'I found a problem', description: "Something didn't work as it should." },
      suggestion: { label: 'I have a suggestion', description: 'An idea to make Cadência better.' },
    },
    loadFailed: 'The feedback page could not be loaded.',
    sendFailed: 'Your feedback could not be sent.',
    loading: 'Loading feedback…',
    kicker: 'HELP CADÊNCIA IMPROVE',
    title: 'How is your experience so far?',
    intro: 'Your report helps us fix details and build ever clearer workouts for people who ride.',
    guideTitle: 'Tell us what you noticed.',
    guideText: 'You can talk about creating your profile, the plan, the workout explanations or any detail of your experience.',
    tips: [
      'Be specific: it makes investigating easier.',
      "Don't include sensitive data or other people's information.",
      'The report is linked to your account so we can organize the feedback.',
    ],
    formTitle: 'Send feedback',
    formIntro: 'It takes less than a minute. Thanks for helping test the app.',
    topic: 'What would you like to talk about?',
    rating: 'How would you rate the experience?',
    ratingScale: '1 = poor · 5 = excellent',
    ratingLabel: 'Experience rating',
    message: 'What would you like to tell us?',
    messagePlaceholder: 'What worked well? What could be clearer?',
    count: (length: number) => `${length}/2000 characters · minimum 10`,
    sent: 'Feedback sent. Thanks for sharing!',
    sending: 'Sending…',
    send: 'Send feedback',
  },
});

export default function FeedbackPage() {
  const t = useMessages(messages);
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
      .catch((caught) => {
        if (caught instanceof ApiError && caught.status === 401) {
          window.location.href = '/entrar';
          return;
        }
        setError(apiErrorMessage(caught, t.loadFailed));
      })
      .finally(() => setLoading(false));
  }, [t]);

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
      setError(caught instanceof Error ? caught.message : t.sendFailed);
    } finally {
      setSending(false);
    }
  }

  if (loading) {
    return <main className="profile-loading"><LoaderCircle className="spin" />{t.loading}</main>;
  }
  if (!user) return <ApiErrorState message={error || t.loadFailed} />;

  return (
    <main className="feedback-shell">
      <AppHeader name={user.display_name} />
      <section className="feedback-content">
        <header className="feedback-heading">
          <p>{t.kicker}</p>
          <h1>{t.title}</h1>
          <span>{t.intro}</span>
        </header>
        <div className="feedback-layout">
          <section className="feedback-guide">
            <span className="feedback-icon"><MessageSquareHeart size={24} /></span>
            <h2>{t.guideTitle}</h2>
            <p>{t.guideText}</p>
            <ul>
              {t.tips.map((tip) => <li key={tip}><CheckCircle2 size={17} /><span>{tip}</span></li>)}
            </ul>
          </section>
          <form className="feedback-form" onSubmit={submit}>
            <h2>{t.formTitle}</h2>
            <p>{t.formIntro}</p>
            <fieldset className="feedback-categories">
              <legend>{t.topic}</legend>
              <div>
                {categories.map((value) => (
                  <label key={value} htmlFor={`feedback-category-${value}`} aria-label={t.categories[value].label} className={category === value ? 'selected' : ''}>
                    <input id={`feedback-category-${value}`} type="radio" name="category" value={value} checked={category === value} onChange={() => setCategory(value)} />
                    <span><strong>{t.categories[value].label}</strong><small>{t.categories[value].description}</small></span>
                  </label>
                ))}
              </div>
            </fieldset>
            <fieldset className="feedback-rating">
              <legend>{t.rating} <small>{t.ratingScale}</small></legend>
              <div aria-label={t.ratingLabel}>
                {[1, 2, 3, 4, 5].map((value) => (
                  <button key={value} type="button" className={rating === value ? 'selected' : ''} aria-pressed={rating === value} onClick={() => setRating(value)}>{value}</button>
                ))}
              </div>
            </fieldset>
            <div className="feedback-message">
              <label htmlFor="feedback-message">{t.message}</label>
              <textarea id="feedback-message" value={message} onChange={(event) => setMessage(event.target.value)} minLength={10} maxLength={2000} placeholder={t.messagePlaceholder} required />
              <small>{t.count(message.length)}</small>
            </div>
            {error && <p className="form-error" role="alert">{error}</p>}
            {submitted && <p className="feedback-success" aria-live="polite"><CheckCircle2 size={17} />{t.sent}</p>}
            <button type="submit" disabled={sending || rating === 0 || message.trim().length < 10}>
              {sending ? <><LoaderCircle className="spin" size={17} />{t.sending}</> : <><Send size={17} />{t.send}</>}
            </button>
          </form>
        </div>
      </section>
    </main>
  );
}
