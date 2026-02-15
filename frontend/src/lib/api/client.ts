/**
 * Centralized API client for communicating with the backend.
 * Automatically handles:
 * - CSRF token extraction from cookies
 * - Request/response serialization
 * - Standard error handling and transformation
 * - Credentials (cookies)
 */

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
			error: 'network_error',
			message: response.statusText || 'Unknown error',
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
}

export async function api<T = unknown>(endpoint: string, options: ApiOptions = {}): Promise<T> {
	const basePath = options.baseUrl ?? '';
	const url = `${basePath}/api/v1${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;

	const csrf = getCsrfToken();
	const customHeaders = options.headers;

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

	const fetchImpl = options.customFetch ?? fetch;

	try {
		const response = await fetchImpl(url, {
			...options,
			credentials: 'include',
			headers
		});

		if (!response.ok) {
			const error = await parseError(response);

			// 401 Unauthorized: session expired or not logged in → redirect to login
			if (response.status === 401 && typeof window !== 'undefined') {
				const { goto } = await import('$app/navigation');
				goto('/login');
				throw new HandledApiException(401, error.error, error.message || 'Unauthorized');
			}

			// Central handling: authorization errors should be surfaced via toast,
			// not rendered inline in table/list UIs.
			if (error.error === 'authorization_failed') {
				const message = error.message || 'You are not authorized to perform this action.';
				if (typeof window !== 'undefined') {
					try {
						const { addToast } = await import('$lib/components/toast.svelte');
						addToast(message, 'error');
					} catch {
						// Ignore toast rendering failures (e.g. during SSR or if component isn't mounted)
					}
				}

				throw new HandledApiException(response.status, error.error, message, error.details);
			}

			throw new ApiException(
				response.status,
				error.error,
				error.message || `HTTP ${response.status}`,
				error.details
			);
		}

		// Handle 204 No Content
		if (response.status === 204) {
			return undefined as T;
		}

		return response.json() as Promise<T>;
	} catch (err) {
		// Re-throw ApiException as-is
		if (err instanceof ApiException) throw err;

		// Network errors
		if (err instanceof TypeError) {
			throw new ApiException(
				0,
				'network_error',
				'Network request failed. Backend may be unavailable.',
				err.message
			);
		}

		// Unknown errors
		throw new ApiException(
			500,
			'unknown_error',
			'An unexpected error occurred',
			err instanceof Error ? err.message : String(err)
		);
	}
}

/**
 * Helper to format error messages for display
 */
export function getErrorMessage(err: unknown): string {
	if (err instanceof ApiException) {
		if (err.message) return err.message;
		if (err.details && typeof err.details === 'object') {
			const entries = Object.entries(err.details as Record<string, string>);
			if (entries.length > 0) {
				return entries.map(([key, value]) => `${key} ${value}`).join(', ');
			}
		}
		return err.error;
	}
	if (err instanceof Error) {
		return err.message;
	}
	return 'An unknown error occurred';
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
	return Object.fromEntries(entries);
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
