import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { SPSController } from '$lib/domain/facility/index.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let spsControllers: SPSController[] = [];

	try {
		const res = await fetch(`${getBackendUrl()}/api/v1/sps-controllers`, {
			headers: {
				Cookie: cookies.get('access_token') ? `access_token=${cookies.get('access_token')}` : ''
			}
		});

		if (res.ok) {
			const data = await res.json();
			spsControllers = data.sps_controllers ?? data ?? [];
		}
	} catch (e) {
		console.error('Failed to load SPS controllers:', e);
	}

	return { spsControllers };
};
