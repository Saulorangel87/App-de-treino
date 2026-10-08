'use client';

import { defineMessages } from '@/lib/i18n';
import { useMessages } from './locale-provider';

const messages = defineMessages({
  pt: { title: 'Não foi possível carregar esta tela.', retry: 'Tentar novamente' },
  en: { title: 'This screen could not be loaded.', retry: 'Try again' },
});

type ApiErrorStateProps = {
  message: string;
};

export function ApiErrorState({ message }: ApiErrorStateProps) {
  const t = useMessages(messages);
  return (
    <main className="api-error-state" role="alert">
      <strong>{t.title}</strong>
      <p>{message}</p>
      <button type="button" onClick={() => window.location.reload()}>
        {t.retry}
      </button>
    </main>
  );
}
