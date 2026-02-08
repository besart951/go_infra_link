import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { getBuilding } from '$lib/infrastructure/api/facility.adapter';

export const load: PageLoad = async ({ params, fetch }) => {
	try {
		const building = await getBuilding(params.id, { customFetch: fetch });
		return { building };
	} catch (e: any) {
		console.error('Failed to load building:', e);
		if (e.status === 404) {
			error(404, 'Building not found');
		}
		error(e.status || 500, 'Failed to load building');
	}
};
