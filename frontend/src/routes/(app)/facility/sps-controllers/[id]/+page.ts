import { error } from '@sveltejs/kit';
import { api } from '$lib/api/client.js';
import type { PageLoad } from './$types';
import type {
  Building,
  ControlCabinet,
  SPSController,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch }) => {
  try {
    const controller = await api<SPSController>(`/facility/sps-controllers/${params.id}`, {
      customFetch: fetch
    });

    const cabinetPromise = controller.control_cabinet_id
      ? api<ControlCabinet>(`/facility/control-cabinets/${controller.control_cabinet_id}`, {
          customFetch: fetch
        })
      : Promise.resolve(null);

    const systemTypesPromise = api<SPSControllerSystemTypeListResponse>(
      `/facility/sps-controller-system-types?page=1&limit=200&sps_controller_id=${controller.id}`,
      { customFetch: fetch }
    );

    const cabinet = await cabinetPromise;
    const buildingPromise = cabinet?.building_id
      ? api<Building>(`/facility/buildings/${cabinet.building_id}`, { customFetch: fetch })
      : Promise.resolve(null);

    const [systemTypesResponse, building] = await Promise.all([
      systemTypesPromise,
      buildingPromise
    ]);

    return {
      controller,
      cabinet,
      building,
      systemTypes: systemTypesResponse.items ?? []
    };
  } catch (e: any) {
    console.error('Failed to load SPS controller detail:', e);
    if (e?.status === 404) {
      error(404, t('facility.sps_controller_not_found'));
    }
    error(e?.status || 500, t('facility.sps_controller_load_failed'));
  }
};
