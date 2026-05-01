import { error } from '@sveltejs/kit';
import { api } from '$lib/api/client.js';
import type { PageLoad } from './$types';
import type {
  Building,
  ControlCabinet,
  FieldDeviceListResponse,
  SPSController,
  SPSControllerSystemType
} from '$lib/domain/facility/index.js';
import { t } from '$lib/i18n/index.js';

export const load: PageLoad = async ({ params, fetch, url }) => {
  try {
    const systemType = await api<SPSControllerSystemType>(
      `/facility/sps-controller-system-types/${params.id}`,
      { customFetch: fetch }
    );

    const controller = await api<SPSController>(
      `/facility/sps-controllers/${systemType.sps_controller_id}`,
      {
        customFetch: fetch
      }
    );

    const cabinetPromise = controller.control_cabinet_id
      ? api<ControlCabinet>(`/facility/control-cabinets/${controller.control_cabinet_id}`, {
          customFetch: fetch
        })
      : Promise.resolve(null);

    const cabinet = await cabinetPromise;
    const buildingPromise = cabinet?.building_id
      ? api<Building>(`/facility/buildings/${cabinet.building_id}`, { customFetch: fetch })
      : Promise.resolve(null);

    const fieldDevicesResponse = await api<FieldDeviceListResponse>(
      `/facility/field-devices?page=1&limit=300&order_by=apparat_nr&order=asc&sps_controller_system_type_id=${systemType.id}`,
      { customFetch: fetch }
    );

    return {
      systemType,
      controller,
      cabinet,
      building: await buildingPromise,
      fieldDevices: fieldDevicesResponse.items ?? [],
      fieldDevicesTotal: fieldDevicesResponse.total ?? 0,
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
