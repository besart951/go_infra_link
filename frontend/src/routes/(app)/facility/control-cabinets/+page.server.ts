import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let controlCabinets: ControlCabinet[] = [];

	try {
		const res = await fetch(`${getBackendUrl()}/api/v1/control-cabinets`, {
			headers: {
				Cookie: cookies.get('access_token') ? `access_token=${cookies.get('access_token')}` : ''
			}
		});

		if (res.ok) {
			const data = await res.json();
			controlCabinets = data.control_cabinets ?? data ?? [];
		}
	} catch (e) {
		console.error('Failed to load control cabinets:', e);
	}

	return { controlCabinets };
};
