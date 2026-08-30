export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

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
    throw new Error(body?.error?.message || 'Não foi possível concluir a solicitação.');
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
