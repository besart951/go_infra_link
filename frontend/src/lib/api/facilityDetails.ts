import { apiClient } from './generated/client.js';
import type { components } from './generated/schema.js';
import {
  facilityDetailCache,
  type FacilityDetail,
  type FacilityDetailKind,
  type FacilityDetailScope
} from '$lib/services/facilityDetailCache.js';

type BuildingDetail =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingDetailResponse'];
type ControlCabinetDetail =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDetailResponse'];
type SPSControllerDetail =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerDetailResponse'];
type SPSControllerSystemTypeDetail =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeDetailResponse'];
type FieldDeviceDetail =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceDetailResponse'];

export type FacilityDetailResponse =
  | BuildingDetail
  | ControlCabinetDetail
  | SPSControllerDetail
  | SPSControllerSystemTypeDetail
  | FieldDeviceDetail;

export type FacilityDetailPatch =
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest'];

function requireData<T>(data: T | undefined): T {
  if (!data) throw new Error('The server returned an empty detail response.');
  return data;
}

export async function loadFacilityDetail(
  kind: FacilityDetailKind,
  id: string,
  scope: FacilityDetailScope = {},
  options: { force?: boolean } = {}
): Promise<FacilityDetailResponse> {
  if (!options.force) {
    const cached = facilityDetailCache.get(kind, id, scope);
    if (cached) return cached as FacilityDetailResponse;
  }

  let detail: FacilityDetailResponse;
  if (scope.projectId) {
    detail = await loadProjectDetail(kind, id, scope.projectId, scope);
  } else {
    detail = await loadGlobalDetail(kind, id, scope);
  }
  facilityDetailCache.set(kind, id, detail as FacilityDetail, scope);
  return detail;
}

function relationQuery(scope: FacilityDetailScope): { page: number; limit: number } {
  return { page: scope.relationPage ?? 1, limit: 12 };
}

async function loadGlobalDetail(
  kind: FacilityDetailKind,
  id: string,
  scope: FacilityDetailScope
): Promise<FacilityDetailResponse> {
  switch (kind) {
    case 'buildings': {
      const { data } = await apiClient.GET('/api/v1/facility/buildings/{id}/detail', {
        params: { path: { id }, query: relationQuery(scope) }
      });
      return requireData(data);
    }
    case 'control-cabinets': {
      const { data } = await apiClient.GET('/api/v1/facility/control-cabinets/{id}/detail', {
        params: { path: { id }, query: relationQuery(scope) }
      });
      return requireData(data);
    }
    case 'sps-controllers': {
      const { data } = await apiClient.GET('/api/v1/facility/sps-controllers/{id}/detail', {
        params: { path: { id }, query: relationQuery(scope) }
      });
      return requireData(data);
    }
    case 'sps-controller-system-types': {
      const { data } = await apiClient.GET(
        '/api/v1/facility/sps-controller-system-types/{id}/detail',
        { params: { path: { id }, query: relationQuery(scope) } }
      );
      return requireData(data);
    }
    case 'field-devices': {
      const { data } = await apiClient.GET('/api/v1/facility/field-devices/{id}/detail', {
        params: { path: { id } }
      });
      return requireData(data);
    }
  }
}

async function loadProjectDetail(
  kind: FacilityDetailKind,
  id: string,
  projectId: string,
  scope: FacilityDetailScope
): Promise<FacilityDetailResponse> {
  switch (kind) {
    case 'buildings': {
      const { data } = await apiClient.GET(
        '/api/v1/projects/{id}/facility/buildings/{buildingId}',
        { params: { path: { id: projectId, buildingId: id }, query: relationQuery(scope) } }
      );
      return requireData(data);
    }
    case 'control-cabinets': {
      const { data } = await apiClient.GET(
        '/api/v1/projects/{id}/facility/control-cabinets/{controlCabinetId}',
        { params: { path: { id: projectId, controlCabinetId: id }, query: relationQuery(scope) } }
      );
      return requireData(data);
    }
    case 'sps-controllers': {
      const { data } = await apiClient.GET(
        '/api/v1/projects/{id}/facility/sps-controllers/{spsControllerId}',
        { params: { path: { id: projectId, spsControllerId: id }, query: relationQuery(scope) } }
      );
      return requireData(data);
    }
    case 'sps-controller-system-types': {
      const { data } = await apiClient.GET(
        '/api/v1/projects/{id}/facility/sps-controller-system-types/{spsControllerSystemTypeId}',
        {
          params: {
            path: { id: projectId, spsControllerSystemTypeId: id },
            query: relationQuery(scope)
          }
        }
      );
      return requireData(data);
    }
    case 'field-devices': {
      const { data } = await apiClient.GET(
        '/api/v1/projects/{id}/facility/field-devices/{fieldDeviceId}',
        { params: { path: { id: projectId, fieldDeviceId: id } } }
      );
      return requireData(data);
    }
  }
}

export async function patchFacilityDetail(
  kind: Exclude<FacilityDetailKind, 'buildings'>,
  id: string,
  patch: FacilityDetailPatch,
  scope: FacilityDetailScope = {}
): Promise<void> {
  if (scope.projectId) {
    await patchProjectDetail(kind, id, scope.projectId, patch);
  } else {
    await patchGlobalDetail(kind, id, patch);
  }
  facilityDetailCache.invalidate(kind, id, scope);
}

export async function patchBuildingDetail(
  id: string,
  patch: components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateBuildingRequest']
): Promise<void> {
  await apiClient.PATCH('/api/v1/facility/buildings/{id}', {
    params: { path: { id } },
    body: patch
  });
  facilityDetailCache.invalidate('buildings', id);
}

async function patchGlobalDetail(
  kind: Exclude<FacilityDetailKind, 'buildings'>,
  id: string,
  patch: FacilityDetailPatch
): Promise<void> {
  switch (kind) {
    case 'control-cabinets':
      await apiClient.PATCH('/api/v1/facility/control-cabinets/{id}', {
        params: { path: { id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest']
      });
      return;
    case 'sps-controllers':
      await apiClient.PATCH('/api/v1/facility/sps-controllers/{id}', {
        params: { path: { id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest']
      });
      return;
    case 'sps-controller-system-types':
      await apiClient.PATCH('/api/v1/facility/sps-controller-system-types/{id}', {
        params: { path: { id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest']
      });
      return;
    case 'field-devices':
      await apiClient.PATCH('/api/v1/facility/field-devices/{id}', {
        params: { path: { id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest']
      });
      return;
  }
}

async function patchProjectDetail(
  kind: Exclude<FacilityDetailKind, 'buildings'>,
  id: string,
  projectId: string,
  patch: FacilityDetailPatch
): Promise<void> {
  switch (kind) {
    case 'control-cabinets':
      await apiClient.PATCH('/api/v1/projects/{id}/facility/control-cabinets/{controlCabinetId}', {
        params: { path: { id: projectId, controlCabinetId: id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateControlCabinetRequest']
      });
      return;
    case 'sps-controllers':
      await apiClient.PATCH('/api/v1/projects/{id}/facility/sps-controllers/{spsControllerId}', {
        params: { path: { id: projectId, spsControllerId: id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerRequest']
      });
      return;
    case 'sps-controller-system-types':
      await apiClient.PATCH(
        '/api/v1/projects/{id}/facility/sps-controller-system-types/{spsControllerSystemTypeId}',
        {
          params: { path: { id: projectId, spsControllerSystemTypeId: id } },
          body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateSPSControllerSystemTypeRequest']
        }
      );
      return;
    case 'field-devices':
      await apiClient.PATCH('/api/v1/projects/{id}/facility/field-devices/{fieldDeviceId}', {
        params: { path: { id: projectId, fieldDeviceId: id } },
        body: patch as components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.UpdateFieldDeviceRequest']
      });
      return;
  }
}
