import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { loadSPSControllerSystemTypeDetailData } from '$lib/application/useCases/facility/loadFacilityDetailData.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch, url }) => {
  try {
    return {
      ...(await loadSPSControllerSystemTypeDetailData(params.id, { customFetch: fetch })),
      editRequested: url.searchParams.get('edit') === '1'
    };
  } catch (e: any) {
    console.error('Failed to load SPS controller system type detail:', e);
    if (e?.status === 404) {
      error(404, t('facility.sps_controller_system_type_not_found'));
    }
    error(e?.status || 500, t('facility.fetch_failed'));
  }
};
