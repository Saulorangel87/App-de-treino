'use client';

type ApiErrorStateProps = {
  message: string;
};

export function ApiErrorState({ message }: ApiErrorStateProps) {
  return (
    <main className="api-error-state" role="alert">
      <strong>Não foi possível carregar esta tela.</strong>
      <p>{message}</p>
      <button type="button" onClick={() => window.location.reload()}>
        Tentar novamente
      </button>
    </main>
  );
}
