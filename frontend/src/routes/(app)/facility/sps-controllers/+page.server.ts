import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { SPSController } from '$lib/domain/facility/index.js';
import { listSPSControllers } from '$lib/infrastructure/api/facility.adapter.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let spsControllers: SPSController[] = [];

	try {
        const accessToken = cookies.get('access_token');
		const csrfToken = cookies.get('csrf_token');
		const cookieHeader = [
			accessToken ? `access_token=${accessToken}` : '',
			csrfToken ? `csrf_token=${csrfToken}` : ''
		]
			.filter(Boolean)
			.join('; ');

		const response = await listSPSControllers(
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
		spsControllers = data.items ?? data.sps_controllers ?? (Array.isArray(data) ? data : []);
	} catch (e) {
		console.error('Failed to load SPS controllers:', e);
	}

	return { spsControllers };
};