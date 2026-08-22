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
import { localizeErrorText } from './errorLocalization.js';
import { fieldErrorsFromApiDetails, parseApiErrorResponse } from './errorResponse.js';
import { resolveFieldError } from './fieldPath.js';
import type { FieldErrorMap } from './errorResponse.js';

export { localizeErrorText, localizeFieldErrorMap } from './errorLocalization.js';
export type { ApiError, FieldErrorMap } from './errorResponse.js';
export { fieldErrorPathMatches } from './fieldPath.js';

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

export function buildHttpErrorRoute(
  status: number,
  fromPath: string,
  message?: string
): string | null {
  const path = getHttpErrorPath(status);
  if (!path) return null;

  const target = new URL(path, 'http://localhost');
  if (fromPath && fromPath !== path) {
    target.searchParams.set('from', fromPath);
  }
  if (message?.trim()) {
    target.searchParams.set('message', message.trim());
  }
  return `${target.pathname}${target.search}`;
}

async function navigateToHttpErrorPage(status: number, message?: string): Promise<boolean> {
  if (typeof window === 'undefined') return false;

  const path = getHttpErrorPath(status);
  if (!path) return false;

  const currentPath = `${window.location.pathname}${window.location.search}${window.location.hash}`;
  if (currentPath === path || currentPath.startsWith(`${path}?`)) {
    return true;
  }

  const route = buildHttpErrorRoute(status, currentPath, message);
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

function prepareHeaders(headersInit?: HeadersInit, includeContentType = false): Headers {
  const headers = new Headers(headersInit);
  if (includeContentType && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const csrf = getCsrfToken();
  if (csrf) {
    headers.set('X-CSRF-Token', csrf);
  }

  return headers;
}

function getInputMethod(input: RequestInfo | URL, options: RequestInit): string {
  if (options.method) return options.method.toUpperCase();
  if (input instanceof Request) return input.method.toUpperCase();
  return 'GET';
}

interface ApiTransportOptions {
  customFetch?: typeof fetch;
  retry?: ApiOptions['retry'];
  retryDelayMs?: number;
}

async function requestWithRetry(
  input: RequestInfo | URL,
  options: RequestInit,
  transport: ApiTransportOptions = {}
): Promise<Response> {
  const fetchImpl = transport.customFetch ?? fetch;
  const method = getInputMethod(input, options);
  const maxRetries = getMaxRetries(method, transport.retry);
  const retryDelayMs = transport.retryDelayMs ?? DEFAULT_RETRY_DELAY_MS;

  for (let attempt = 0; attempt <= maxRetries; attempt += 1) {
    try {
      const response = await fetchImpl(input, options);
      if (attempt < maxRetries && isRetryableResponse(response)) {
        reportApiRetry(attempt + 1, maxRetries);
        await wait(retryDelay(attempt + 1, response, retryDelayMs), options.signal ?? undefined);
        continue;
      }
      return response;
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw err;
      }

      if (isNetworkError(err)) {
        if (attempt < maxRetries) {
          reportApiRetry(attempt + 1, maxRetries);
          await wait(retryDelay(attempt + 1, undefined, retryDelayMs), options.signal ?? undefined);
          continue;
        }

        reportApiFailure();
        throw new ApiException(0, 'network_error', t('errors.network_request_failed'), err.message);
      }

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

export async function throwApiResponse(
  response: Response,
  skipHttpErrorNavigation = false
): Promise<never> {
  if (isRetryableResponse(response)) {
    reportApiFailure();
  }

  const error = await parseApiErrorResponse(response);

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

  const message = localizeErrorText(
    error.message || response.statusText || `HTTP ${response.status}`
  );

  if (!skipHttpErrorNavigation && (await navigateToHttpErrorPage(response.status, message))) {
    throw new HandledApiException(response.status, error.error, message, error.details);
  }

  throw new ApiException(response.status, error.error, message, error.details);
}

export async function assertApiSuccess(
  response: Response,
  skipHttpErrorNavigation = false
): Promise<Response> {
  if (!response.ok) {
    return throwApiResponse(response, skipHttpErrorNavigation);
  }

  reportApiSuccess();
  return response;
}

/**
 * Fetch adapter for generated OpenAPI clients. It preserves the same-origin
 * cookie, CSRF and safe-retry policy used by the hand-written API wrapper.
 */
export async function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const request = new Request(input, init);
  const preparedRequest = new Request(request, {
    credentials: 'include',
    headers: prepareHeaders(
      request.headers,
      request.body !== null && !(init?.body instanceof FormData)
    )
  });

  return requestWithRetry(preparedRequest, {}, {});
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

  const response = await requestWithRetry(
    url,
    {
      ...fetchOptions,
      credentials: 'include',
      headers: prepareHeaders(fetchOptions.headers, true)
    },
    { customFetch, retry, retryDelayMs }
  );

  await assertApiSuccess(response, skipHttpErrorNavigation);
  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
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
  return fieldErrorsFromApiDetails(err.details);
}

/**
 * Resolve a field error by checking both prefixed and raw keys.
 */
export function getFieldError(
  errors: FieldErrorMap,
  field: string,
  prefixes: string[] = []
): string | undefined {
  return resolveFieldError(errors, field, prefixes);
}

/**
 * Check if error is a specific type
 */
export function isApiError(err: unknown, errorCode?: string): boolean {
  if (!(err instanceof ApiException)) return false;
  return errorCode ? err.error === errorCode : true;
}
