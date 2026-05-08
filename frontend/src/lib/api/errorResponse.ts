import { localizeFieldErrorMap } from './errorLocalization.js';

export interface ApiError {
  error: string;
  code?: string;
  message?: string;
  details?: unknown;
  localized_key?: string;
  field_errors?: unknown;
  request_id?: string;
  status?: number;
}

export type FieldErrorMap = Record<string, string>;

export async function parseApiErrorResponse(response: Response): Promise<ApiError> {
  try {
    const body = await response.json();
    const code = typeof body.code === 'string' && body.code ? body.code : body.error;
    return {
      error: code || 'unknown_error',
      code,
      message: body.message || body.localized_key || response.statusText,
      localized_key: body.localized_key,
      field_errors: body.field_errors,
      details: body.field_errors || body.fields || body.details,
      request_id: body.request_id,
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

export function fieldErrorsFromApiDetails(details: unknown): FieldErrorMap {
  if (!details || typeof details !== 'object') return {};

  if (Array.isArray(details)) {
    const mapped = details.reduce<FieldErrorMap>((acc, item) => {
      if (!item || typeof item !== 'object') return acc;
      const fieldError = item as Record<string, unknown>;
      const path = typeof fieldError.path === 'string' ? fieldError.path : '';
      if (!path) return acc;

      const localizedKey =
        typeof fieldError.localized_key === 'string' && fieldError.localized_key
          ? fieldError.localized_key
          : '';
      const rawMessage =
        typeof fieldError.message === 'string' && fieldError.message ? fieldError.message : '';
      const code = typeof fieldError.code === 'string' && fieldError.code ? fieldError.code : '';
      const message = localizedKey.startsWith('validation.')
        ? rawMessage || localizedKey
        : localizedKey || rawMessage || code;

      if (message) {
        acc[path] = message;
      }
      return acc;
    }, {});
    return localizeFieldErrorMap(mapped);
  }

  const entries = Object.entries(details as Record<string, unknown>)
    .filter(([, value]) => typeof value === 'string')
    .map(([key, value]) => [key, value as string]);
  return localizeFieldErrorMap(Object.fromEntries(entries));
}
