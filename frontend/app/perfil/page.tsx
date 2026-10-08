'use client';

import { Check, LoaderCircle, MailCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { AppHeader } from '@/components/app-header';
import { useMessages } from '@/components/locale-provider';
import { defineMessages } from '@/lib/i18n';
import { STEP_ICONS } from './profile-model';
import { useProfileForm } from './use-profile-form';
import { ProfileStep1 } from './step-basics';
import { ProfileStep2 } from './step-safety';
import { ProfileStep3 } from './step-goals';
import { ProfileStep4 } from './step-routine';

const messages = defineMessages({
  pt: {
    loading: 'Carregando seu perfil…',
    confirmEmail: 'Confirme seu e-mail antes de gerar ou ativar um plano.',
    sentTo: (email: string) => `Enviamos um link para ${email}.`,
    sending: 'Enviando…',
    resend: 'Reenviar link',
    progress: 'Progresso do perfil',
    of: 'de 04',
    reviewLater: 'Você poderá revisar todas essas informações quando sua rotina mudar.',
  },
  en: {
    loading: 'Loading your profile…',
    confirmEmail: 'Confirm your email before generating or activating a plan.',
    sentTo: (email: string) => `We sent a link to ${email}.`,
    sending: 'Sending…',
    resend: 'Resend link',
    progress: 'Profile progress',
    of: 'of 04',
    reviewLater: 'You can review all of this information whenever your routine changes.',
  },
});

export default function ProfilePage() {
  const t = useMessages(messages);
  const form = useProfileForm();
  const {
    AsideIcon,
    completed,
    copy,
    loading,
    resendVerification,
    sendingVerification,
    step,
    user,
    verificationMessage,
  } = form;

  if (loading)
    return (
      <main className="profile-loading">
        <LoaderCircle className="spin" />
        {t.loading}
      </main>
    );

  return (
    <main className="profile-shell">
      <AppHeader name={user?.display_name} />
      <section className="profile-content">
        {user && !user.email_verified && (
          <section className="email-verification-banner">
            <MailCheck size={19} />
            <div>
              <strong>{t.confirmEmail}</strong>
              <p>{verificationMessage || t.sentTo(user.email)}</p>
            </div>
            <Button
              type="button"
              variant="outline"
              disabled={sendingVerification}
              onClick={resendVerification}
            >
              {sendingVerification ? t.sending : t.resend}
            </Button>
          </section>
        )}
        <nav className="onboarding-progress" aria-label={t.progress}>
          {STEP_ICONS.map((_, index) => (
            <span key={index} className={index + 1 <= step ? 'active' : ''}>
              <i>
                {index + 1 < step || completed ? (
                  <Check size={11} />
                ) : (
                  index + 1
                )}
              </i>
            </span>
          ))}
        </nav>
        <div className="profile-heading">
          <div>
            <p>{copy.kicker}</p>
            <h1>{copy.title}</h1>
            <span>{copy.description}</span>
          </div>
          <div className="step-indicator">
            <strong>0{step}</strong>
            <span>{t.of}</span>
          </div>
        </div>
        <div className="profile-layout">
          {step === 1 && <ProfileStep1 form={form} />}

          {step === 2 && <ProfileStep2 form={form} />}

          {step === 3 && <ProfileStep3 form={form} />}

          {step === 4 && <ProfileStep4 form={form} />}

          <aside className="profile-aside">
            <div className="aside-icon">
              <AsideIcon size={18} />
            </div>
            <h2>{copy.asideTitle}</h2>
            <p>{copy.aside}</p>
            <hr />
            <span>{t.reviewLater}</span>
          </aside>
        </div>
      </section>
    </main>
  );
}
