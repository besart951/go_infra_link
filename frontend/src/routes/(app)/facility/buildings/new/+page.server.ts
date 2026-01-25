import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types.js';
import { getBackendUrl } from '$lib/server/backend.js';
import { createBuilding } from '$lib/infrastructure/api/facility.adapter.js';

interface FormErrors {
	iws_code?: string;
	building_group?: string;
	form?: string;
}

interface FormValues {
	iws_code?: string;
	building_group?: string;
}

export const actions: Actions = {
	default: async ({ request, fetch, cookies }) => {
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

			await createBuilding(
				{
					iws_code: iws_code!,
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

			// Success - redirect to buildings list
		} catch (e) {
			console.error('Failed to create building:', e);
			// Check if e is ApiException
			const message = (e as any).message || 'An unexpected error occurred';
			return fail(500, {
				errors: { form: message } as FormErrors,
				values: { iws_code, building_group } as FormValues
			});
		}

		redirect(303, '/facility/buildings');
	}
};
