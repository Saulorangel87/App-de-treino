import { currentLocale, HTML_LANG } from './i18n';

export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const DEFAULT_TIMEOUT_MS = 20_000;
const DEFAULT_GET_RETRIES = 2;
const RETRY_BASE_DELAY_MS = 300;
const RETRYABLE_STATUS = new Set([502, 503, 504]);
export const PUBLIC_PATHS = ['/entrar', '/esqueci-minha-senha', '/redefinir-senha', '/verificar-email', '/privacidade', '/termos'];

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export type ApiRequestOptions = RequestInit & {
  /** Tempo máximo de cada tentativa. Padrão: 20 s. */
  timeoutMs?: number;
  /** Novas tentativas após falha de rede ou 502/503/504. Só vale para GET; padrão: 2. */
  retries?: number;
};

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export function isTimeout(error: unknown): boolean {
  return error instanceof ApiError && error.status === 0;
}

const text = {
  pt: {
    offline: 'Não foi possível conectar à API. Verifique sua conexão e tente novamente.',
    timeout: 'A solicitação demorou demais para responder. Tente novamente.',
    failed: 'Não foi possível concluir a solicitação.',
    downloadFailed: 'Não foi possível baixar o arquivo.',
  },
  en: {
    offline: 'Could not reach the server. Check your connection and try again.',
    timeout: 'The request took too long to respond. Please try again.',
    failed: 'The request could not be completed.',
    downloadFailed: 'The file could not be downloaded.',
  },
};

function messages() {
  return text[currentLocale()];
}

/** A API responde mensagens e conteúdo dos treinos no idioma escolhido no app, não no do navegador. */
function withLanguage(headers: Headers): Headers {
  headers.set('Accept-Language', HTML_LANG[currentLocale()]);
  return headers;
}

export function apiErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof ApiError) return error.message;
  // Cada navegador usa uma mensagem diferente ("Failed to fetch", "Load failed",
  // "NetworkError..."); toda falha de rede do fetch é um TypeError.
  if (error instanceof TypeError) {
    return messages().offline;
  }
  if (error instanceof Error) return error.message;
  return fallback;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function redirectToLoginOnExpiredSession(path: string) {
  if (typeof window === 'undefined') return;
  // Rotas de autenticação respondem 401 para credenciais inválidas; isso não é
  // sessão expirada e a própria tela mostra o erro.
  if (path.startsWith('/v1/auth/')) return;
  if (PUBLIC_PATHS.includes(window.location.pathname)) return;
  window.location.href = '/entrar';
}

async function fetchWithTimeout(url: string, init: RequestInit, timeoutMs: number): Promise<Response> {
  const controller = new AbortController();
  let timedOut = false;
  const timer = setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeoutMs);
  const callerSignal = init.signal;
  const forwardAbort = () => controller.abort();
  if (callerSignal) {
    if (callerSignal.aborted) controller.abort();
    else callerSignal.addEventListener('abort', forwardAbort, { once: true });
  }
  try {
    return await fetch(url, { ...init, signal: controller.signal });
  } catch (error) {
    if (timedOut) {
      throw new ApiError(messages().timeout, 0);
    }
    throw error;
  } finally {
    clearTimeout(timer);
    callerSignal?.removeEventListener('abort', forwardAbort);
  }
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const { timeoutMs = DEFAULT_TIMEOUT_MS, retries, ...init } = options;
  const method = (init.method || 'GET').toUpperCase();
  // Só requisições idempotentes são repetidas; um POST repetido poderia duplicar efeitos.
  const maxRetries = method === 'GET' ? (retries ?? DEFAULT_GET_RETRIES) : 0;

  const headers = withLanguage(new Headers(init.headers));
  // Um corpo FormData (upload de arquivo) precisa que o navegador defina o
  // Content-Type sozinho, com o boundary do multipart; forçar application/json
  // quebraria o envio.
  if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  const requestInit: RequestInit = { ...init, headers, credentials: 'include' };

  let attempt = 0;
  for (;;) {
    try {
      const response = await fetchWithTimeout(`${API_URL}${path}`, requestInit, timeoutMs);
      if (RETRYABLE_STATUS.has(response.status) && attempt < maxRetries) {
        attempt += 1;
        await sleep(RETRY_BASE_DELAY_MS * 2 ** (attempt - 1));
        continue;
      }
      if (!response.ok) {
        const body = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
        if (response.status === 401) redirectToLoginOnExpiredSession(path);
        throw new ApiError(body?.error?.message || messages().failed, response.status);
      }
      if (response.status === 204) return undefined as T;
      return (await response.json()) as T;
    } catch (error) {
      const retryable = error instanceof TypeError || isTimeout(error);
      if (retryable && attempt < maxRetries && !init.signal?.aborted) {
        attempt += 1;
        await sleep(RETRY_BASE_DELAY_MS * 2 ** (attempt - 1));
        continue;
      }
      throw error;
    }
  }
}

const DOWNLOAD_TIMEOUT_MS = 60_000;

/** Baixa um arquivo gerado pela API (por exemplo, a exportação de dados) e o salva no aparelho. */
export async function apiDownload(path: string, filename: string): Promise<void> {
  const response = await fetchWithTimeout(`${API_URL}${path}`, { credentials: 'include', headers: withLanguage(new Headers()) }, DOWNLOAD_TIMEOUT_MS);
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    if (response.status === 401) redirectToLoginOnExpiredSession(path);
    throw new ApiError(body?.error?.message || messages().downloadFailed, response.status);
  }
  const url = URL.createObjectURL(await response.blob());
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
