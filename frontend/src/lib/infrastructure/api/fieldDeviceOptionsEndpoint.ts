import { api, type ApiOptions } from '$lib/api/client.js';
import type {
  AvailableApparatNumbersResponse,
  FieldDeviceOptions
} from '$lib/domain/facility/index.js';

export function getFieldDeviceOptions(options?: ApiOptions): Promise<FieldDeviceOptions> {
  return api<FieldDeviceOptions>('/facility/field-devices/options', options);
}

export function getFieldDeviceOptionsForProject(
  projectId: string,
  options?: ApiOptions
): Promise<FieldDeviceOptions> {
  return api<FieldDeviceOptions>(`/projects/${projectId}/field-device-options`, options);
}

export function getAvailableApparatNumbers(
  spsControllerSystemTypeId: string,
  apparatId: string,
  systemPartId?: string,
  options?: ApiOptions
): Promise<AvailableApparatNumbersResponse> {
  const searchParams = new URLSearchParams();
  searchParams.set('sps_controller_system_type_id', spsControllerSystemTypeId);
  searchParams.set('apparat_id', apparatId);
  if (systemPartId) {
    searchParams.set('system_part_id', systemPartId);
  }

  return api<AvailableApparatNumbersResponse>(
    `/facility/field-devices/available-apparat-nr?${searchParams.toString()}`,
    options
  );
}
