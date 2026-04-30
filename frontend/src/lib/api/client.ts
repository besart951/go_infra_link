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

const FIELD_LABEL_KEYS: Record<string, string> = {
  apparat: 'facility.apparat',
  apparat_id: 'facility.apparat',
  apparat_nr: 'field_device.table.apparat_nr',
  bacnet_objects: 'facility.bacnet_object',
  bacnetobject: 'facility.bacnet_object',
  bmk: 'field_device.table.bmk',
  building_group: 'facility.building_group',
  building_id: 'facility.building',
  controlcabinet: 'facility.control_cabinet',
  control_cabinet_id: 'facility.control_cabinet',
  control_cabinet_nr: 'facility.forms.control_cabinet.number_label',
  description: 'common.description',
  device_name: 'facility.device_name',
  channel: 'notifications.preferences.channel_title',
  code: 'notifications.preferences.email.code_label',
  event_key: 'notifications.rules.event_key',
  fielddevice: 'facility.field_device',
  frequency: 'notifications.preferences.frequency_title',
  from_email: 'notifications.form.from_email',
  ga_device: 'facility.ga_device',
  gateway: 'facility.forms.sps_controller.gateway_label',
  host: 'notifications.form.host',
  ip_address: 'facility.ip_address',
  iws_code: 'facility.iws_code',
  name: 'common.name',
  notification_email: 'notifications.preferences.email.label',
  objectdata: 'facility.object_data',
  object_data_id: 'facility.object_data',
  password: 'notifications.form.password',
  phase_id: 'projects.settings.phase',
  port: 'notifications.form.port',
  project_id: 'notifications.rules.project_id',
  reply_to: 'notifications.form.reply_to',
  recipient_role: 'notifications.rules.role',
  recipient_team_id: 'notifications.rules.team_id',
  recipient_type: 'notifications.rules.recipient_type',
  recipient_user_ids: 'notifications.rules.user_ids',
  resource_id: 'notifications.rules.resource_id',
  resource_type: 'notifications.rules.resource_type',
  spscontroller: 'facility.sps_controller',
  specification: 'facility.specifications',
  subnet: 'facility.forms.sps_controller.subnet_label',
  system_part: 'facility.system_part',
  system_part_id: 'facility.system_part',
  system_types: 'facility.forms.sps_controller.system_types_title',
  text_fix: 'field_device.table.text_fix',
  to: 'notifications.test.to',
  username: 'notifications.form.username',
  vlan: 'facility.forms.sps_controller.vlan_label'
};

const SCOPE_LABEL_KEYS: Record<string, string> = {
  building: 'facility.building',
  control_cabinet: 'facility.control_cabinet',
  'control cabinet': 'facility.control_cabinet',
  ip_address: 'facility.ip_address',
  vlan: 'facility.forms.sps_controller.vlan_label'
};

const DIRECT_MESSAGE_KEYS: Record<string, string> = {
  'Bad Request': 'errors.bad_request',
  Conflict: 'errors.conflict',
  'Failed to load notification count': 'notifications.errors.notification_count_load_failed',
  'Failed to load notification preference': 'notifications.errors.preference_load_failed',
  'Failed to load notifications': 'notifications.errors.notifications_load_failed',
  'Failed to load notification rules': 'notifications.errors.rules_load_failed',
  'Failed to load SMTP settings': 'notifications.errors.smtp_settings_load_failed',
  'Failed to mark notification read': 'notifications.errors.mark_read_failed',
  'Failed to mark notifications read': 'notifications.errors.mark_all_read_failed',
  'Failed to save notification preference': 'notifications.errors.preference_save_failed',
  'Failed to save notification rule': 'notifications.errors.rule_save_failed',
  'Failed to save SMTP settings': 'notifications.errors.smtp_settings_save_failed',
  'Failed to verify notification email': 'notifications.errors.email_verification_failed',
  Forbidden: 'errors.forbidden',
  'Internal Server Error': 'errors.internal_server_error',
  'Notification not found': 'notifications.errors.notification_not_found',
  'Notification rule not found': 'notifications.errors.rule_not_found',
  'Not Found': 'errors.not_found',
  'SMTP settings not configured': 'notifications.errors.smtp_settings_not_configured',
  'Unknown error': 'errors.unknown_error',
  Unauthorized: 'errors.unauthorized',
  authorization_failed: 'errors.unauthorized',
  email_verification_failed: 'notifications.errors.email_verification_failed',
  fetch_failed: 'errors.fetch_failed',
  'field device is required': 'field_device.multi_create.validation.field_device_required',
  'notification provider disabled': 'notifications.errors.smtp_disabled',
  'notification provider not configured': 'notifications.errors.smtp_settings_not_configured',
  'Failed to delete notification rule': 'notifications.errors.rule_delete_failed',
  'no available ga_device for control cabinet': 'facility.no_available_ga_device',
  'object_data_id and bacnet_objects are mutually exclusive': 'facility.mutually_exclusive_error',
  'one or more parent entities (SPS controller, apparat, system part) not found':
    'field_device.multi_create.validation.parents_not_found',
  'apparat_nr is required': 'field_device.multi_create.validation.apparat_nr_required',
  'apparat_nr must be between 1 and 99': 'field_device.validation.apparat_nr_range',
  'apparatnummer ist bereits vergeben': 'field_device.multi_create.validation.apparat_nr_used',
  referenced_entity_in_use: 'facility.referenced_entity_in_use',
  smtp_test_failed: 'notifications.errors.smtp_test_failed',
  update_failed: 'errors.update_failed',
  validation_error: 'errors.validation_error'
};

function translateIfExists(key: string, params?: Record<string, string | number>): string | null {
  const translated = t(key, params);
  return translated !== key ? translated : null;
}

function extractFieldSegment(fieldPath?: string): string | undefined {
  if (!fieldPath) return undefined;
  const segments = fieldPath.split('.').filter(Boolean);
  return segments.length > 0 ? segments[segments.length - 1] : fieldPath;
}

function humanizeFieldName(field?: string): string {
  if (!field) return '';
  return field.replaceAll('_', ' ');
}

function getFieldLabel(fieldPath?: string): string {
  const field = extractFieldSegment(fieldPath);
  if (!field) return '';
  const translationKey = FIELD_LABEL_KEYS[field];
  if (translationKey) {
    return t(translationKey);
  }
  return humanizeFieldName(field);
}

function getScopeLabel(scope: string): string {
  const normalized = scope.trim().toLowerCase();
  const translationKey = SCOPE_LABEL_KEYS[normalized];
  if (translationKey) {
    return t(translationKey);
  }
  return humanizeFieldName(normalized);
}

export function localizeErrorText(message: string, fieldPath?: string): string {
  const trimmed = message.trim();
  if (!trimmed) return trimmed;

  const translatedByKey = translateIfExists(trimmed);
  if (translatedByKey) {
    return translatedByKey;
  }

  const translatedByErrorKey = translateIfExists(`errors.${trimmed}`);
  if (translatedByErrorKey) {
    return translatedByErrorKey;
  }

  const directTranslationKey = DIRECT_MESSAGE_KEYS[trimmed];
  if (directTranslationKey) {
    return t(directTranslationKey);
  }

  if (trimmed === 'is required') {
    return t('validation.required', { field: getFieldLabel(fieldPath) });
  }

  if (trimmed === 'must be a valid email') {
    return t('validation.email_invalid', { field: getFieldLabel(fieldPath) });
  }

  if (trimmed === 'invalid') {
    return t('validation.invalid', { field: getFieldLabel(fieldPath) });
  }

  if (
    /^(smtp dial|smtp tls dial|smtp client|smtp starttls|smtp auth|smtp mail from|smtp rcpt to|smtp data|smtp write|smtp close data|smtp quit):/i.test(
      trimmed
    )
  ) {
    return t('notifications.errors.smtp_delivery_failed');
  }

  if (/^(decode secret|decrypt secret):/i.test(trimmed)) {
    return t('notifications.errors.smtp_secret_failed');
  }

  let match = trimmed.match(/^([a-z0-9_.-]+) is required$/i);
  if (match) {
    return t('validation.required', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) is required when auth_mode is plain$/i);
  if (match) {
    return t('validation.required_when_plain_auth', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^max (\d+)$/i);
  if (match) {
    return t('validation.max_generic', { field: getFieldLabel(fieldPath), max: match[1] });
  }

  match = trimmed.match(/^min (\d+)$/i);
  if (match) {
    return t('validation.min_generic', { field: getFieldLabel(fieldPath), min: match[1] });
  }

  match = trimmed.match(/^length (\d+)$/i);
  if (match) {
    return t('validation.exact_length', { field: getFieldLabel(fieldPath), length: match[1] });
  }

  match = trimmed.match(/^must be one of: (.+)$/i);
  if (match) {
    return t('validation.one_of', { field: getFieldLabel(fieldPath), options: match[1] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be (\d+) characters or less$/i);
  if (match) {
    return t('validation.max_length', { field: getFieldLabel(match[1]), max: match[2] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a valid IPv4 address$/i);
  if (match) {
    return t('validation.valid_ipv4', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a valid IPv4 subnet mask$/i);
  if (match) {
    return t('validation.valid_ipv4_subnet', { field: getFieldLabel(match[1]) });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be a number between (\d+) and (\d+)$/i);
  if (match) {
    return t('validation.number_between', {
      field: getFieldLabel(match[1]),
      min: match[2],
      max: match[3]
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be between (\d+) and (\d+)$/i);
  if (match) {
    return t('validation.range', { field: getFieldLabel(match[1]), min: match[2], max: match[3] });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be exactly (\d+) uppercase letters \(A-Z\)$/i);
  if (match) {
    return t('validation.exact_uppercase_letters', {
      field: getFieldLabel(match[1]),
      count: match[2]
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be unique within the (.+)$/i);
  if (match) {
    return t('validation.unique_within', {
      field: getFieldLabel(match[1]),
      scope: getScopeLabel(match[2])
    });
  }

  match = trimmed.match(/^([a-z0-9_.-]+) must be unique per (.+)$/i);
  if (match) {
    return t('validation.unique_per', {
      field: getFieldLabel(match[1]),
      scope: getScopeLabel(match[2])
    });
  }

  return trimmed;
}

export function localizeFieldErrorMap(errors: Record<string, string>): FieldErrorMap {
  return Object.fromEntries(
    Object.entries(errors).map(([field, value]) => [field, localizeErrorText(value, field)])
  );
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
