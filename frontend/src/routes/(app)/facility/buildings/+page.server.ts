import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { Building } from '$lib/domain/facility/index.js';
import { listBuildings } from '$lib/infrastructure/api/facility.adapter.js';

export const load: PageServerLoad = async ({ fetch, cookies }) => {
	let buildings: Building[] = [];

	try {
		const accessToken = cookies.get('access_token');
		const csrfToken = cookies.get('csrf_token');
		const cookieHeader = [
			accessToken ? `access_token=${accessToken}` : '',
			csrfToken ? `csrf_token=${csrfToken}` : ''
		].filter(Boolean).join('; ');

		const response = await listBuildings(
			{ limit: 100 }, // Fetch reasonable amount
			{
				baseUrl: getBackendUrl(),
				customFetch: fetch,
				headers: {
					Cookie: cookieHeader
				}
			}
		);

		// Handle various response formats from backend
		// The API adapter returns BuildingListResponse which extends Pagination<Building> (has 'items')
		// But in case the backend returns something else or we're dealing with raw response:
		const data = response as any;
		buildings = data.items ?? data.buildings ?? (Array.isArray(data) ? data : []);
	} catch (e) {
		console.error('Failed to load buildings:', e);
	}

	return { buildings };
};
