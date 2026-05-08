import { api, type ApiOptions } from '$lib/api/client.js';
import type {
  Building,
  ControlCabinet,
  SPSController,
  SPSControllerListResponse,
  SPSControllerSystemType,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';

interface LoadOptions {
  customFetch?: ApiOptions['customFetch'];
}

function apiOptions(options: LoadOptions): Pick<ApiOptions, 'customFetch'> {
  return { customFetch: options.customFetch };
}

export async function loadBuildingDetailData(id: string, options: LoadOptions = {}) {
  const building = await api<Building>(`/facility/buildings/${id}`, apiOptions(options));
  return { building };
}

export async function loadControlCabinetDetailData(id: string, options: LoadOptions = {}) {
  const requestOptions = apiOptions(options);
  const cabinet = await api<ControlCabinet>(`/facility/control-cabinets/${id}`, requestOptions);

  const buildingPromise = cabinet.building_id
    ? api<Building>(`/facility/buildings/${cabinet.building_id}`, requestOptions)
    : Promise.resolve(null);

  const spsResponse = await api<SPSControllerListResponse>(
    `/facility/sps-controllers?page=1&limit=200&control_cabinet_id=${cabinet.id}`,
    requestOptions
  );

  const controllers = spsResponse.items ?? [];
  const systemTypesByController: Record<string, SPSControllerSystemType[]> = {};

  if (controllers.length > 0) {
    const controllerIds = controllers.map((controller: SPSController) => controller.id).join('|');
    const limit = Math.max(1000, controllers.length * 200);
    const systemTypeResponse = await api<SPSControllerSystemTypeListResponse>(
      `/facility/sps-controller-system-types?page=1&limit=${limit}&sps_controller_id=${controllerIds}`,
      requestOptions
    );

    for (const systemType of systemTypeResponse.items ?? []) {
      if (!systemTypesByController[systemType.sps_controller_id]) {
        systemTypesByController[systemType.sps_controller_id] = [];
      }
      systemTypesByController[systemType.sps_controller_id].push(systemType);
    }
  }

  return {
    cabinet,
    building: await buildingPromise,
    spsControllers: controllers,
    systemTypesByController
  };
}

export async function loadSPSControllerDetailData(id: string, options: LoadOptions = {}) {
  const requestOptions = apiOptions(options);
  const controller = await api<SPSController>(`/facility/sps-controllers/${id}`, requestOptions);

  const cabinetPromise = controller.control_cabinet_id
    ? api<ControlCabinet>(
        `/facility/control-cabinets/${controller.control_cabinet_id}`,
        requestOptions
      )
    : Promise.resolve(null);

  const systemTypesPromise = api<SPSControllerSystemTypeListResponse>(
    `/facility/sps-controller-system-types?page=1&limit=200&sps_controller_id=${controller.id}`,
    requestOptions
  );

  const cabinet = await cabinetPromise;
  const buildingPromise = cabinet?.building_id
    ? api<Building>(`/facility/buildings/${cabinet.building_id}`, requestOptions)
    : Promise.resolve(null);

  const [systemTypesResponse, building] = await Promise.all([systemTypesPromise, buildingPromise]);

  return {
    controller,
    cabinet,
    building,
    systemTypes: systemTypesResponse.items ?? []
  };
}

export async function loadSPSControllerSystemTypeDetailData(id: string, options: LoadOptions = {}) {
  const requestOptions = apiOptions(options);
  const systemType = await api<SPSControllerSystemType>(
    `/facility/sps-controller-system-types/${id}`,
    requestOptions
  );

  const controller = await api<SPSController>(
    `/facility/sps-controllers/${systemType.sps_controller_id}`,
    requestOptions
  );

  const cabinet = controller.control_cabinet_id
    ? await api<ControlCabinet>(
        `/facility/control-cabinets/${controller.control_cabinet_id}`,
        requestOptions
      )
    : null;

  const building = cabinet?.building_id
    ? await api<Building>(`/facility/buildings/${cabinet.building_id}`, requestOptions)
    : null;

  return {
    systemType,
    controller,
    cabinet,
    building
  };
}
