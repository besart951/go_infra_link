import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import { getSetCookieValues } from '$lib/server/set-cookie.js';

interface AuthResponse {
	csrf_token: string;
	access_token_expires_at: string;
	refresh_token_expires_at: string;
}

function secondsUntil(isoDate: string): number | undefined {
	const d = new Date(isoDate);
	if (Number.isNaN(d.getTime())) return undefined;
	const seconds = Math.floor((d.getTime() - Date.now()) / 1000);
	return seconds > 0 ? seconds : undefined;
}

export const actions: Actions = {
	default: async ({ request, fetch, cookies, url }) => {
		const form = await request.formData();
		const email = String(form.get('email') ?? '').trim();
		const password = String(form.get('password') ?? '');

		if (!email || !password) {
			return fail(400, { error: 'missing_fields', message: 'Email and password are required' });
		}

		let res: Response;
		try {
			res = await fetch(`${getBackendUrl()}/api/v1/auth/login`, {
				method: 'POST',
				headers: {
					'content-type': 'application/json'
				},
				body: JSON.stringify({ email, password })
			});
		} catch (err) {
			// Network error (backend down, DNS issue, connection refused, etc.)
			const msg = err instanceof Error ? err.message : 'Network request failed';
			return fail(503, { error: 'service_unavailable', message: `Backend unavailable: ${msg}` });
		}

		if (!res.ok) {
			let error = 'login_failed';
			let message = 'Login failed';
			try {
				const body = (await res.json()) as { error?: string; message?: string };
				error = body.error ?? error;
				message = body.message ?? message;
			} catch {
				// Ignore JSON parse errors, use defaults
			}
			return fail(res.status, { error, message });
		}

		const body = (await res.json()) as AuthResponse;
		const setCookies = getSetCookieValues(res);

		const accessToken = setCookies.access_token;
		const refreshToken = setCookies.refresh_token;

		if (!accessToken || !refreshToken) {
			return fail(500, {
				error: 'missing_auth_cookies',
				message: 'Backend did not return auth cookies'
			});
		}

		const secure = url.protocol === 'https:';
		const accessMaxAge = secondsUntil(body.access_token_expires_at);
		const refreshMaxAge = secondsUntil(body.refresh_token_expires_at);

		cookies.set('access_token', accessToken, {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure,
			maxAge: accessMaxAge
		});
		cookies.set('refresh_token', refreshToken, {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure,
			maxAge: refreshMaxAge
		});
		cookies.set('csrf_token', body.csrf_token, {
			path: '/',
			httpOnly: false,
			sameSite: 'lax',
			secure,
			maxAge: 60 * 60 * 24
		});

		throw redirect(303, '/');
	}
};
