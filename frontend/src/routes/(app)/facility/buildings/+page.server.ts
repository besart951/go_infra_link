import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { Building } from '$lib/domain/facility/index.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let buildings: Building[] = [];

	try {
		const res = await fetch(`${getBackendUrl()}/api/v1/buildings`, {
			headers: {
				Cookie: cookies.get('access_token') ? `access_token=${cookies.get('access_token')}` : ''
			}
		});

		if (res.ok) {
			const data = await res.json();
			buildings = data.buildings ?? data ?? [];
		}
	} catch (e) {
		console.error('Failed to load buildings:', e);
	}

	return { buildings };
};
