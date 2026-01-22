import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const { pathname } = event.url;

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

	if (!isInternal) {
		if (!event.locals.authenticated && !isLoginRoute) {
			throw redirect(303, '/login');
		}
		if (event.locals.authenticated && isLoginRoute) {
			throw redirect(303, '/');
		}
	}

	return resolve(event);
};
