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
			details: body.details,
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
export async function api<T = unknown>(
	endpoint: string,
	options: RequestInit = {}
): Promise<T> {
	const url = `/api/v1${endpoint.startsWith('/') ? endpoint : '/' + endpoint}`;

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

	try {
		const response = await fetch(url, {
			...options,
			credentials: 'include',
			headers
		});

		if (!response.ok) {
			const error = await parseError(response);
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
		return err.message || err.error;
	}
	if (err instanceof Error) {
		return err.message;
	}
	return 'An unknown error occurred';
}

/**
 * Check if error is a specific type
 */
export function isApiError(err: unknown, errorCode?: string): boolean {
	if (!(err instanceof ApiException)) return false;
	return errorCode ? err.error === errorCode : true;
}
