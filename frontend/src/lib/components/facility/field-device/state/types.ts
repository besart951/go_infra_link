import type { TableFilterRecord } from '$lib/state/table/contracts.js';

export interface FieldDeviceFilters extends TableFilterRecord {
  buildingId?: string;
  controlCabinetId?: string;
  spsControllerId?: string;
  spsControllerSystemTypeId?: string;
  projectId?: string;
}

export type ProjectIdInput = string | undefined | (() => string | undefined);

export interface SharedFieldDeviceDraftState {
  devices: Array<{
    device_id: string;
    changed_fields: string[];
    field_values?: Record<string, unknown>;
  }>;
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
  pageSize?: number;
  sharedFieldDeviceEditors?: () => SharedFieldDeviceEditorsByDevice;
  onSharedFieldDeviceStateChange?: (state: SharedFieldDeviceDraftState) => void;
  onFieldDevicesSaved?: (deviceIds: string[]) => void;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}
