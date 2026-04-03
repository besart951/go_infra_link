import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { api } from '$lib/api/client.js';
import type { Building } from '$lib/domain/facility/index.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch }) => {
  try {
    const building = await api<Building>(`/facility/buildings/${params.id}`, {
      customFetch: fetch
    });
    return { building };
  } catch (e: any) {
    console.error('Failed to load building:', e);
    if (e.status === 404) {
      error(404, t('facility.building_not_found'));
    }
    error(e.status || 500, t('facility.building_load_failed'));
  }
};
