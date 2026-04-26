import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import type { FieldDevice, SPSController } from '$lib/domain/facility/index.js';

export interface FieldDeviceFilters extends TableFilterRecord {
  buildingId?: string;
  controlCabinetId?: string;
  spsControllerId?: string;
  spsControllerSystemTypeId?: string;
  projectId?: string;
}

export type ProjectIdInput = string | undefined | (() => string | undefined);
export type PageSizeInput = number | undefined | (() => number | undefined);

export interface SharedFieldDeviceDraftState {
  devices: Array<{
    device_id: string;
    changed_fields: string[];
    field_values?: Record<string, unknown>;
  }>;
}

export interface FieldDeviceRefreshRequest {
  key: string | number;
  devices?: FieldDevice[];
  deviceIds?: string[];
  spsControllers?: SPSController[];
  spsControllerIds?: string[];
}

export interface SharedFieldDeviceEditor {
  userId: string;
  firstName: string;
  lastName: string;
  changedFields: string[];
  fieldValues?: Record<string, unknown>;
  updatedAt: string;
}

export type SharedFieldDeviceEditorsByDevice = Record<string, SharedFieldDeviceEditor[]>;

export interface FieldDeviceStateProps {
  projectId?: ProjectIdInput;
  pageSize?: PageSizeInput;
  sharedFieldDeviceEditors?: () => SharedFieldDeviceEditorsByDevice;
  onSharedFieldDeviceStateChange?: (state: SharedFieldDeviceDraftState) => void;
  onFieldDevicesSaved?: (devices: FieldDevice[]) => void;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}

export function resolvePageSize(pageSize?: PageSizeInput): number | undefined {
  if (typeof pageSize === 'function') {
    return pageSize();
  }

  return pageSize;
}
