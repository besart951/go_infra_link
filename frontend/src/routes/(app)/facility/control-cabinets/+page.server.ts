import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import { listControlCabinets } from '$lib/infrastructure/api/facility.adapter.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let controlCabinets: ControlCabinet[] = [];

	try {
        const accessToken = cookies.get('access_token');
		const csrfToken = cookies.get('csrf_token');
		const cookieHeader = [
			accessToken ? `access_token=${accessToken}` : '',
			csrfToken ? `csrf_token=${csrfToken}` : ''
		]
			.filter(Boolean)
			.join('; ');

		const response = await listControlCabinets(
            { limit: 100 },
            {
                baseUrl: getBackendUrl(),
                customFetch: fetch,
                headers: {
                    Cookie: cookieHeader
                }
            }
        );

		const data = response as any;
		controlCabinets = data.items ?? data.control_cabinets ?? (Array.isArray(data) ? data : []);
	} catch (e) {
		console.error('Failed to load control cabinets:', e);
	}

	return { controlCabinets };
};