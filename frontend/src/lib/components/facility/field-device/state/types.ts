import type { TableFilterRecord } from '$lib/state/table/contracts.js';

export interface FieldDeviceFilters extends TableFilterRecord {
  buildingId?: string;
  controlCabinetId?: string;
  spsControllerId?: string;
  spsControllerSystemTypeId?: string;
  projectId?: string;
}

export type ProjectIdInput = string | undefined | (() => string | undefined);

export interface FieldDeviceStateProps {
  projectId?: ProjectIdInput;
  pageSize?: number;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}
