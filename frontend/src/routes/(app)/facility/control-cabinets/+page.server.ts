import type { PageServerLoad } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import { listControlCabinets } from '$lib/infrastructure/api/facility.adapter.js';

export const load: PageServerLoad = async ({ fetch, cookies, url }) => {
	const page = Number(url.searchParams.get('page')) || 1;
	const limit = Number(url.searchParams.get('limit')) || 10;
	const search = url.searchParams.get('search') || '';

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
			{ page, limit, search },
			{
				baseUrl: getBackendUrl(),
				customFetch: fetch,
				headers: {
					Cookie: cookieHeader
				}
			}
		);

		return {
			controlCabinets: response.items || [],
			total: response.total || 0,
			page: response.page || page,
			total_pages: response.total_pages || 1,
			limit
		};
	} catch (e) {
		console.error('Failed to load control cabinets:', e);
		return {
			controlCabinets: [],
			total: 0,
			page: 1,
			totalPages: 1,
			limit
		};
	}
};