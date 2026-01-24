import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { FieldDevice } from '$lib/domain/facility/index.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let fieldDevices: FieldDevice[] = [];

	try {
		const res = await fetch(`${getBackendUrl()}/api/v1/field-devices`, {
			headers: {
				Cookie: cookies.get('access_token') ? `access_token=${cookies.get('access_token')}` : ''
			}
		});

		if (res.ok) {
			const data = await res.json();
			fieldDevices = data.field_devices ?? data ?? [];
		}
	} catch (e) {
		console.error('Failed to load field devices:', e);
	}

	return { fieldDevices };
};
