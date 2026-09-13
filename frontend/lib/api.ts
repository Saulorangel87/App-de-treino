export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export function apiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) return error.message;
  if (error instanceof TypeError && error.message === 'Failed to fetch') {
    return 'Não foi possível conectar à API. Verifique se ela está em execução e tente novamente.';
  }
  if (error instanceof Error) return error.message;
  return fallback;
}

export async function apiRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null) as { error?: { message?: string } } | null;
    throw new ApiError(body?.error?.message || 'Não foi possível concluir a solicitação.', response.status);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
