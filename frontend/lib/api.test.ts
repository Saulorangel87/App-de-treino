import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, apiErrorMessage, apiRequest } from './api';

function jsonResponse(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('apiRequest', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('fetch', fetchMock);
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
    fetchMock.mockReset();
  });

  it('returns parsed JSON and sends credentials', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(200, { ok: true }));
    await expect(apiRequest<{ ok: boolean }>('/v1/me')).resolves.toEqual({ ok: true });
    expect(fetchMock.mock.calls[0][1]).toMatchObject({ credentials: 'include' });
  });

  it('sends the app language, not the browser one, so the API answers in it', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(200, {}));
    await apiRequest('/v1/me');
    expect((fetchMock.mock.calls[0][1].headers as Headers).get('Accept-Language')).toBe('pt-BR');

    vi.stubGlobal('document', { cookie: 'cadencia_lang=en' });
    fetchMock.mockResolvedValueOnce(jsonResponse(200, {}));
    await apiRequest('/v1/me');
    expect((fetchMock.mock.calls[1][1].headers as Headers).get('Accept-Language')).toBe('en');
  });

  it('returns undefined for 204 responses', async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));
    await expect(apiRequest('/v1/auth/logout', { method: 'POST' })).resolves.toBeUndefined();
  });

  it('retries GET requests after a 503 and then succeeds', async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(503)).mockResolvedValueOnce(jsonResponse(200, { plan: null }));
    const promise = apiRequest('/v1/plans/current');
    await vi.runAllTimersAsync();
    await expect(promise).resolves.toEqual({ plan: null });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('retries GET requests after a network failure', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('Load failed')).mockResolvedValueOnce(jsonResponse(200, { ok: 1 }));
    const promise = apiRequest('/v1/me');
    await vi.runAllTimersAsync();
    await expect(promise).resolves.toEqual({ ok: 1 });
  });

  it('never retries non-idempotent requests', async () => {
    fetchMock.mockResolvedValue(jsonResponse(503, { error: { message: 'indisponível' } }));
    const promise = apiRequest('/v1/plans/generate', { method: 'POST' });
    const assertion = expect(promise).rejects.toMatchObject({ status: 503, message: 'indisponível' });
    await vi.runAllTimersAsync();
    await assertion;
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it('gives up after the configured retries', async () => {
    fetchMock.mockResolvedValue(jsonResponse(502));
    const promise = apiRequest('/v1/me', { retries: 1 });
    const assertion = expect(promise).rejects.toBeInstanceOf(ApiError);
    await vi.runAllTimersAsync();
    await assertion;
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('turns a hung request into a timeout ApiError', async () => {
    fetchMock.mockImplementation(
      (_url: string, init: RequestInit) =>
        new Promise((_resolve, reject) => {
          init.signal?.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError')));
        }),
    );
    const promise = apiRequest('/v1/plans/generate', { method: 'POST', timeoutMs: 1000 });
    const assertion = expect(promise).rejects.toMatchObject({ status: 0 });
    await vi.advanceTimersByTimeAsync(1001);
    await assertion;
  });

  it('does not treat auth 401 as an expired session', async () => {
    const location = { href: '/plano', pathname: '/plano' };
    vi.stubGlobal('window', { location });
    fetchMock.mockResolvedValueOnce(jsonResponse(401, { error: { message: 'E-mail ou senha inválidos.' } }));
    await expect(apiRequest('/v1/auth/login', { method: 'POST' })).rejects.toMatchObject({ status: 401 });
    expect(location.href).toBe('/plano');
  });

  it('redirects to the login page when the session expires on a protected route', async () => {
    const location = { href: '/plano', pathname: '/plano' };
    vi.stubGlobal('window', { location });
    fetchMock.mockResolvedValueOnce(jsonResponse(401, { error: { message: 'Faça login para continuar.' } }));
    await expect(apiRequest('/v1/workouts/1/complete', { method: 'POST' })).rejects.toMatchObject({ status: 401 });
    expect(location.href).toBe('/entrar');
  });

  it('does not redirect when already on a public page', async () => {
    const location = { href: '/entrar', pathname: '/entrar' };
    vi.stubGlobal('window', { location });
    fetchMock.mockResolvedValueOnce(jsonResponse(401));
    await expect(apiRequest('/v1/me')).rejects.toMatchObject({ status: 401 });
    expect(location.href).toBe('/entrar');
  });
});

describe('apiErrorMessage', () => {
  it('describes any browser network failure', () => {
    expect(apiErrorMessage(new TypeError('Load failed'), 'x')).toMatch(/conectar/);
    expect(apiErrorMessage(new TypeError('Failed to fetch'), 'x')).toMatch(/conectar/);
  });
  it('uses the API message when present', () => {
    expect(apiErrorMessage(new ApiError('Mensagem', 400), 'x')).toBe('Mensagem');
  });
});
