import type { LayoutServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';

export const load: LayoutServerLoad = async ({ fetch }) => {
	let backendAvailable = true;
	try {
		const res = await fetch(`${getBackendUrl()}/health`, {
			method: 'GET'
		});
		backendAvailable = res.ok;
	} catch {
		backendAvailable = false;
	}

	return { backendAvailable };
};
