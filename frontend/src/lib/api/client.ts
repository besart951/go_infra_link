/**
 * Centralized API client for communicating with the backend.
 * Automatically handles:
 * - CSRF token extraction from cookies
 * - Request/response serialization
 * - Standard error handling and transformation
 * - Credentials (cookies)
 */

import { t } from '$lib/i18n/index.js';
import { reportApiFailure, reportApiRetry, reportApiSuccess } from '$lib/stores/network.js';
import { localizeErrorText, localizeFieldErrorMap } from './errorLocalization.js';

export { localizeErrorText, localizeFieldErrorMap } from './errorLocalization.js';

export interface ApiError {
  error: string;
  message?: string;
  details?: unknown;
  status?: number;
}

export type FieldErrorMap = Record<string, string>;

export class ApiException extends Error {
  constructor(
    public status: number,
    public error: string,
    public message: string,
    public details?: unknown
  ) {
    super(message || error);
    this.name = 'ApiException';
  }
}

/**
 * ApiException that has already been surfaced to the user (e.g. via toast).
 * UI layers can choose to not render it again.
 */
export class HandledApiException extends ApiException {
  public readonly handled = true;
  constructor(status: number, error: string, message: string, details?: unknown) {
    super(status, error, message, details);
    this.name = 'HandledApiException';
  }
}

export function getHttpErrorPath(status: number): string | null {
  if (status === 403) return '/errors/403';
  if (status === 404) return '/errors/404';
  return null;
}

export function buildHttpErrorRoute(status: number, fromPath: string): string | null {
  const path = getHttpErrorPath(status);
  if (!path) return null;

  const target = new URL(path, 'http://localhost');
  if (fromPath && fromPath !== path) {
    target.searchParams.set('from', fromPath);
  }
  return `${target.pathname}${target.search}`;
}

async function navigateToHttpErrorPage(status: number): Promise<boolean> {
  if (typeof window === 'undefined') return false;

  const path = getHttpErrorPath(status);
  if (!path) return false;

  const currentPath = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  if (currentPath === path || currentPath.startsWith(`${path}?`)) {
    return true;
  }

  const route = buildHttpErrorRoute(status, currentPath);
  if (!route) return false;

  const { goto } = await import('$app/navigation');
  await goto(route, { replaceState: true });
  return true;
}

/**
 * Extract CSRF token from cookies
 */
function getCsrfToken(): string | undefined {
  if (typeof document === 'undefined') return undefined;
  const m = document.cookie.match(new RegExp(`(?:^|; )csrf_token=([^;]*)`));
  return m ? decodeURIComponent(m[1]) : undefined;
}

/**
 * Parse backend error response
 */
async function parseError(response: Response): Promise<ApiError> {
  try {
    const body = await response.json();
    return {
      error: body.error || 'unknown_error',
      message: body.message || response.statusText,
      details: body.details || body.fields,
      status: response.status
    };
  } catch {
    return {
      error: 'unknown_error',
      message: response.statusText || 'unknown_error',
      status: response.status
    };
  }
}

/**
 * Main API client function
 * Usage:
 *   const user = await api<User>('/users/me');
 *   await api('/teams', { method: 'POST', body: {...} });
 */
export interface ApiOptions extends RequestInit {
  customFetch?: typeof fetch;
  baseUrl?: string;
  skipHttpErrorNavigation?: boolean;
  retry?: number | false;
  retryDelayMs?: number;
}

const RETRYABLE_STATUS_CODES = new Set([408, 429, 502, 503, 504]);
const SAFE_RETRY_METHODS = new Set(['GET', 'HEAD']);
const DEFAULT_SAFE_RETRIES = 2;
const DEFAULT_RETRY_DELAY_MS = 400;
const MAX_RETRY_AFTER_DELAY_MS = 5000;

function getRequestMethod(options: RequestInit): string {
  return (options.method ?? 'GET').toUpperCase();
}

function getMaxRetries(method: string, retry: ApiOptions['retry']): number {
  if (!SAFE_RETRY_METHODS.has(method) || retry === false) {
    return 0;
  }

  if (typeof retry === 'number') {
    return Math.max(0, Math.floor(retry));
  }

  return DEFAULT_SAFE_RETRIES;
}

function retryDelay(
  attempt: number,
  response?: Response,
  baseDelayMs = DEFAULT_RETRY_DELAY_MS
): number {
  const retryAfter = response?.headers.get('Retry-After');
  if (retryAfter) {
    const seconds = Number(retryAfter);
    if (Number.isFinite(seconds) && seconds >= 0) {
      return Math.min(seconds * 1000, MAX_RETRY_AFTER_DELAY_MS);
    }

    const retryDate = Date.parse(retryAfter);
    if (!Number.isNaN(retryDate)) {
      return Math.min(Math.max(retryDate - Date.now(), 0), MAX_RETRY_AFTER_DELAY_MS);
    }
  }

  if (baseDelayMs <= 0) {
    return 0;
  }

  const jitter = Math.floor(Math.random() * 100);
  return Math.max(0, baseDelayMs) * 2 ** Math.max(0, attempt - 1) + jitter;
}

function wait(ms: number, signal?: AbortSignal): Promise<void> {
  if (ms <= 0) return Promise.resolve();
  if (signal?.aborted) {
    return Promise.reject(new DOMException('Aborted', 'AbortError'));
  }

  return new Promise((resolve, reject) => {
    const timeout = setTimeout(resolve, ms);
    signal?.addEventListener(
      'abort',
      () => {
        clearTimeout(timeout);
        reject(new DOMException('Aborted', 'AbortError'));
      },
      { once: true }
    );
  });
}

function isRetryableResponse(response: Response): boolean {
  return RETRYABLE_STATUS_CODES.has(response.status);
}

function isNetworkError(err: unknown): err is TypeError {
  return err instanceof TypeError;
}

export async function api<T = unknown>(endpoint: string, options: ApiOptions = {}): Promise<T> {
  const {
    baseUrl,
    customFetch,
    skipHttpErrorNavigation = false,
    retry,
    retryDelayMs = DEFAULT_RETRY_DELAY_MS,
    ...fetchOptions
  } = options;

  const basePath = baseUrl ?? '';
  const url = `${basePath}/api/v1${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;

  const csrf = getCsrfToken();
  const customHeaders = fetchOptions.headers;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };

  if (customHeaders) {
    if (customHeaders instanceof Headers) {
      for (const [key, value] of customHeaders.entries()) {
        headers[key] = value;
      }
    } else if (Array.isArray(customHeaders)) {
      for (const header of customHeaders) {
        headers[header[0]] = header[1];
      }
    } else {
      for (const key in customHeaders) {
        headers[key] = customHeaders[key] as string;
      }
    }
  }

  // Add CSRF token if available
  if (csrf) {
    headers['X-CSRF-Token'] = csrf;
  }

  const fetchImpl = customFetch ?? fetch;
  const method = getRequestMethod(fetchOptions);
  const maxRetries = getMaxRetries(method, retry);
  const signal = fetchOptions.signal ?? undefined;

  for (let attempt = 0; attempt <= maxRetries; attempt += 1) {
    try {
      const response = await fetchImpl(url, {
        ...fetchOptions,
        credentials: 'include',
        headers
      });

      if (!response.ok) {
        if (attempt < maxRetries && isRetryableResponse(response)) {
          reportApiRetry(attempt + 1, maxRetries);
          await wait(retryDelay(attempt + 1, response, retryDelayMs), signal);
          continue;
        }

        if (isRetryableResponse(response)) {
          reportApiFailure();
        }

        const error = await parseError(response);

        // 401 Unauthorized: session expired or not logged in → redirect to login
        if (response.status === 401 && typeof window !== 'undefined') {
          const { goto } = await import('$app/navigation');
          goto('/login');
          throw new HandledApiException(
            401,
            error.error,
            localizeErrorText(error.message || 'Unauthorized'),
            error.details
          );
        }

        if (!skipHttpErrorNavigation && (await navigateToHttpErrorPage(response.status))) {
          throw new HandledApiException(
            response.status,
            error.error,
            localizeErrorText(error.message || response.statusText || `HTTP ${response.status}`),
            error.details
          );
        }

        throw new ApiException(
          response.status,
          error.error,
          localizeErrorText(error.message || `HTTP ${response.status}`),
          error.details
        );
      }

      reportApiSuccess();

      // Handle 204 No Content
      if (response.status === 204) {
        return undefined as T;
      }

      return response.json() as Promise<T>;
    } catch (err) {
      // Preserve abort semantics so callers can silently ignore cancellations
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw err;
      }

      // Re-throw ApiException as-is
      if (err instanceof ApiException) throw err;

      // Network errors
      if (isNetworkError(err)) {
        if (attempt < maxRetries) {
          reportApiRetry(attempt + 1, maxRetries);
          await wait(retryDelay(attempt + 1, undefined, retryDelayMs), signal);
          continue;
        }

        reportApiFailure();
        throw new ApiException(0, 'network_error', t('errors.network_request_failed'), err.message);
      }

      // Unknown errors
      throw new ApiException(
        500,
        'unknown_error',
        t('errors.unexpected_error'),
        err instanceof Error ? err.message : String(err)
      );
    }
  }

  throw new ApiException(500, 'unknown_error', t('errors.unexpected_error'));
}

/**
 * Helper to format error messages for display
 */
export function getErrorMessage(err: unknown): string {
  if (err instanceof ApiException) {
    const localizedDetails = getFieldErrors(err);
    if (Object.keys(localizedDetails).length > 0 && err.error === 'validation_error') {
      const first = Object.entries(localizedDetails)[0];
      if (first) {
        return first[1];
      }
    }

    if (err.message) return localizeErrorText(err.message);

    if (err.details && typeof err.details === 'object') {
      const entries = Object.entries(localizedDetails);
      if (entries.length > 0) {
        return entries.map(([, value]) => value).join(', ');
      }
    }
    return localizeErrorText(err.error);
  }
  if (err instanceof Error) {
    return localizeErrorText(err.message);
  }
  return t('errors.unknown_error');
}

/**
 * Extract field-level validation errors from API exceptions.
 */
export function getFieldErrors(err: unknown): FieldErrorMap {
  if (!(err instanceof ApiException)) return {};
  if (!err.details || typeof err.details !== 'object') return {};
  const entries = Object.entries(err.details as Record<string, unknown>)
    .filter(([, value]) => typeof value === 'string')
    .map(([key, value]) => [key, value as string]);
  return localizeFieldErrorMap(Object.fromEntries(entries));
}

/**
 * Resolve a field error by checking both prefixed and raw keys.
 */
export function getFieldError(
  errors: FieldErrorMap,
  field: string,
  prefixes: string[] = []
): string | undefined {
  for (const prefix of prefixes) {
    const key = `${prefix}.${field}`;
    if (errors[key]) return errors[key];
  }
  return errors[field];
}

/**
 * Check if error is a specific type
 */
export function isApiError(err: unknown, errorCode?: string): boolean {
  if (!(err instanceof ApiException)) return false;
  return errorCode ? err.error === errorCode : true;
}
