import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { ProjectCapabilities } from '$lib/domain/project/capabilities.js';
import type { EntityChangeEvent } from '../../shared/entityRefresh.js';

export interface ControlCabinetFilters extends TableFilterRecord {
  buildingIds?: string;
}

export type ProjectIdInput = string | undefined | (() => string | undefined);
export type ProjectCapabilitiesInput =
  | ProjectCapabilities
  | undefined
  | (() => ProjectCapabilities | undefined);

export interface ControlCabinetStateProps {
  projectId?: ProjectIdInput;
  projectCapabilities?: ProjectCapabilitiesInput;
  pageSize?: number;
  onChanged?: (event?: EntityChangeEvent<ControlCabinet>) => void;
}

export function toProjectCapabilitiesResolver(
  capabilities?: ProjectCapabilitiesInput
): () => ProjectCapabilities | undefined {
  if (typeof capabilities === 'function') {
    return capabilities;
  }

  return () => capabilities;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}
