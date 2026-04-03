import { error } from '@sveltejs/kit';
import { api } from '$lib/api/client.js';
import type { PageLoad } from './$types';
import type {
  Building,
  ControlCabinet,
  SPSController,
  SPSControllerListResponse,
  SPSControllerSystemType,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch }) => {
  try {
    const cabinet = await api<ControlCabinet>(`/facility/control-cabinets/${params.id}`, {
      customFetch: fetch
    });

    const buildingPromise = cabinet.building_id
      ? api<Building>(`/facility/buildings/${cabinet.building_id}`, { customFetch: fetch })
      : Promise.resolve(null);

    const spsResponse = await api<SPSControllerListResponse>(
      `/facility/sps-controllers?page=1&limit=200&control_cabinet_id=${cabinet.id}`,
      { customFetch: fetch }
    );

    const systemTypeEntries = await Promise.all(
      (spsResponse.items ?? []).map(async (controller: SPSController) => {
        const response = await api<SPSControllerSystemTypeListResponse>(
          `/facility/sps-controller-system-types?page=1&limit=200&sps_controller_id=${controller.id}`,
          { customFetch: fetch }
        );
        return [controller.id, response.items ?? []] as const;
      })
    );

    const systemTypesByController: Record<string, SPSControllerSystemType[]> =
      Object.fromEntries(systemTypeEntries);

    return {
      cabinet,
      building: await buildingPromise,
      spsControllers: spsResponse.items ?? [],
      systemTypesByController
    };
  } catch (e: any) {
    console.error('Failed to load control cabinet detail:', e);
    if (e?.status === 404) {
      error(404, t('facility.control_cabinet_not_found'));
    }
    error(e?.status || 500, t('facility.control_cabinet_load_failed'));
  }
};
