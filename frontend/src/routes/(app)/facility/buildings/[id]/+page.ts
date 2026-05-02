import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { loadBuildingDetailData } from '$lib/application/useCases/facility/loadFacilityDetailData.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch }) => {
  try {
    return await loadBuildingDetailData(params.id, { customFetch: fetch });
  } catch (e: any) {
    console.error('Failed to load building:', e);
    if (e.status === 404) {
      error(404, t('facility.building_not_found'));
    }
    error(e.status || 500, t('facility.building_load_failed'));
  }
};
