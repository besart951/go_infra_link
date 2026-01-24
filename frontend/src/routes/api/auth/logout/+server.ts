import { redirect, type RequestEvent } from '@sveltejs/kit';
import type { RequestHandler } from './$types.js';

export const GET: RequestHandler = async ({ cookies }: RequestEvent) => {
	cookies.delete('access_token', { path: '/' });
	cookies.delete('refresh_token', { path: '/' });
	cookies.delete('csrf_token', { path: '/' });
	throw redirect(307, '/login');
};
