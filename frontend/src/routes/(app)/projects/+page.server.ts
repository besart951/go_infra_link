import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { Project } from '$lib/domain/project/index.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let projects: Project[] = [];

	try {
		const res = await fetch(`${getBackendUrl()}/api/v1/projects`, {
			headers: {
				Cookie: cookies.get('access_token') ? `access_token=${cookies.get('access_token')}` : ''
			}
		});

		if (res.ok) {
			const data = await res.json();
			projects = data.projects ?? data ?? [];
		}
	} catch (e) {
		console.error('Failed to load projects:', e);
	}

	return { projects };
};
