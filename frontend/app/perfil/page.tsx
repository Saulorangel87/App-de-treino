'use client';

import { Check, LoaderCircle, MailCheck } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { AppHeader } from '@/components/app-header';
import { stepCopy } from './profile-model';
import { useProfileForm } from './use-profile-form';
import { ProfileStep1 } from './step-basics';
import { ProfileStep2 } from './step-safety';
import { ProfileStep3 } from './step-goals';
import { ProfileStep4 } from './step-routine';

export default function ProfilePage() {
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
        Carregando seu perfil…
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
              <strong>
                Confirme seu e-mail antes de gerar ou ativar um plano.
              </strong>
              <p>
                {verificationMessage || `Enviamos um link para ${user.email}.`}
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              disabled={sendingVerification}
              onClick={resendVerification}
            >
              {sendingVerification ? 'Enviando…' : 'Reenviar link'}
            </Button>
          </section>
        )}
        <nav className="onboarding-progress" aria-label="Progresso do perfil">
          {stepCopy.map((_, index) => (
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
            <span>de 04</span>
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
            <span>
              Você poderá revisar todas essas informações quando sua rotina
              mudar.
            </span>
          </aside>
        </div>
      </section>
    </main>
  );
}
