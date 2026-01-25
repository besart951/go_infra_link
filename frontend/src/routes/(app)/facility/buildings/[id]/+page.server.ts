import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import type { Building } from '$lib/domain/facility/index.js';
import { getBuilding, updateBuilding, deleteBuilding } from '$lib/infrastructure/api/facility.adapter.js';

interface FormErrors {
	iws_code?: string;
	building_group?: string;
	form?: string;
}

interface FormValues {
	iws_code?: string;
	building_group?: string;
}

export const load: PageServerLoad = async ({ params, fetch, cookies }) => {
	try {
		const accessToken = cookies.get('access_token');
		const csrfToken = cookies.get('csrf_token');
		const cookieHeader = [
			accessToken ? `access_token=${accessToken}` : '',
			csrfToken ? `csrf_token=${csrfToken}` : ''
		].filter(Boolean).join('; ');

		const building = await getBuilding(params.id, {
			baseUrl: getBackendUrl(),
			customFetch: fetch,
			headers: {
				Cookie: cookieHeader
			}
		});

		return { building };
	} catch (e: any) {
		console.error('Failed to load building:', e);
		if (e.status === 404) {
			error(404, 'Building not found');
		}
		error(e.status || 500, 'Failed to load building');
	}
};

export const actions: Actions = {
	update: async ({ params, request, fetch, cookies }) => {
		const formData = await request.formData();
		const iws_code = formData.get('iws_code')?.toString().trim();
		const building_group = formData.get('building_group')?.toString().trim();

		// Validation
		const errors: FormErrors = {};

		if (!iws_code) {
			errors.iws_code = 'IWS Code is required';
		}

		if (!building_group) {
			errors.building_group = 'Building Group is required';
		} else if (isNaN(Number(building_group))) {
			errors.building_group = 'Building Group must be a number';
		}

		if (Object.keys(errors).length > 0) {
			return fail(400, {
				errors,
				values: { iws_code, building_group } as FormValues
			});
		}

		try {
			const accessToken = cookies.get('access_token');
			const csrfToken = cookies.get('csrf_token');
			const cookieHeader = [
				accessToken ? `access_token=${accessToken}` : '',
				csrfToken ? `csrf_token=${csrfToken}` : ''
			].filter(Boolean).join('; ');

			await updateBuilding(
				params.id,
				{
					iws_code,
					building_group: Number(building_group)
				},
				{
					baseUrl: getBackendUrl(),
					customFetch: fetch,
					headers: {
						Cookie: cookieHeader,
						'X-CSRF-Token': csrfToken || ''
					}
				}
			);

			return { success: true };
		} catch (e: any) {
			console.error('Failed to update building:', e);
			return fail(e.status || 500, {
				errors: { form: e.message || 'An unexpected error occurred' } as FormErrors,
				values: { iws_code, building_group } as FormValues
			});
		}
	},

	delete: async ({ params, fetch, cookies }) => {
		try {
			const accessToken = cookies.get('access_token');
			const csrfToken = cookies.get('csrf_token');
			const cookieHeader = [
				accessToken ? `access_token=${accessToken}` : '',
				csrfToken ? `csrf_token=${csrfToken}` : ''
			].filter(Boolean).join('; ');

			await deleteBuilding(params.id, {
				baseUrl: getBackendUrl(),
				customFetch: fetch,
				headers: {
					Cookie: cookieHeader,
					'X-CSRF-Token': csrfToken || ''
				}
			});
		} catch (e: any) {
			console.error('Failed to delete building:', e);
			return fail(e.status || 500, {
				errors: { form: e.message || 'An unexpected error occurred' } as FormErrors
			});
		}

		redirect(303, '/facility/buildings');
	}
};
