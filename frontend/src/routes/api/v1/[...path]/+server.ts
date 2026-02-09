import type { RequestHandler } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';

function cloneHeaders(src: Headers, options: { skipSetCookie?: boolean } = {}): Headers {
	const h = new Headers();
	src.forEach((value, key) => {
		const lower = key.toLowerCase();
		// Let fetch set these appropriately.
		if (lower === 'host' || lower === 'content-length') return;
		if (options.skipSetCookie && lower === 'set-cookie') return;
		h.set(key, value);
	});
	return h;
}

async function proxy({ request, params }: Parameters<RequestHandler>[0]): Promise<Response> {
	const url = new URL(request.url);
	const targetUrl = `${getBackendUrl()}/api/v1/${params.path}${url.search}`;

	const headers = cloneHeaders(request.headers);
	// Ensure we don't send compressed responses back to node and then re-stream.
	headers.delete('accept-encoding');

	const init: RequestInit = {
		method: request.method,
		headers,
		body:
			request.method === 'GET' || request.method === 'HEAD'
				? undefined
				: await request.arrayBuffer(),
		redirect: 'manual'
	};

	const upstream = await fetch(targetUrl, init);

	const resHeaders = cloneHeaders(upstream.headers, { skipSetCookie: true });
	const setCookies = upstream.headers.getSetCookie?.() ?? [];
	for (const cookie of setCookies) {
		resHeaders.append('set-cookie', cookie);
	}
	return new Response(upstream.body, {
		status: upstream.status,
		statusText: upstream.statusText,
		headers: resHeaders
	});
}

export const GET: RequestHandler = proxy;
export const POST: RequestHandler = proxy;
export const PUT: RequestHandler = proxy;
export const PATCH: RequestHandler = proxy;
export const DELETE: RequestHandler = proxy;
export const OPTIONS: RequestHandler = proxy;
