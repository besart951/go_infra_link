import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';
import { building } from '$app/environment';

export const handle: Handle = async ({ event, resolve }) => {
	// Bypass auth checks during build time (for adapter-static fallback generation)
	if (building) {
		return await resolve(event);
	}

	const { pathname } = event.url;

	// Never redirect API requests to HTML login pages.
	// API callers expect JSON (e.g. 401), not a 303 -> /login HTML document.
	const isApiRoute = pathname.startsWith('/api/');

	// Allow internal assets/routes without auth checks.
	const isInternal =
		pathname.startsWith('/_app') ||
		pathname.startsWith('/favicon') ||
		pathname.startsWith('/robots') ||
		pathname.startsWith('/sitemap') ||
		pathname.startsWith('/manifest') ||
		pathname.startsWith('/icons');

	const isLoginRoute = pathname === '/login';

	const accessToken = event.cookies.get('access_token');
	event.locals.authenticated = Boolean(accessToken);

	if (!isInternal && !isApiRoute) {
		if (!event.locals.authenticated && !isLoginRoute) {
			throw redirect(303, '/login');
		}
		if (event.locals.authenticated && isLoginRoute) {
			throw redirect(303, '/');
		}
	}

	return resolve(event);
};
