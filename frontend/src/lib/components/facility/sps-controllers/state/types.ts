import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import type { SPSController } from '$lib/domain/facility/index.js';
import type { ProjectCapabilities } from '$lib/domain/project/capabilities.js';
import type { EntityChangeEvent } from '../../shared/entityRefresh.js';

export interface SPSControllerFilters extends TableFilterRecord {
  controlCabinetIds?: string;
}

export type ProjectIdInput = string | undefined | (() => string | undefined);
export type ProjectCapabilitiesInput =
  | ProjectCapabilities
  | undefined
  | (() => ProjectCapabilities | undefined);
export type RefreshKeyInput = string | number | undefined | (() => string | number | undefined);

export interface SPSControllerStateProps {
  projectId?: ProjectIdInput;
  projectCapabilities?: ProjectCapabilitiesInput;
  pageSize?: number;
  controlCabinetRefreshKey?: RefreshKeyInput;
  onChanged?: (event?: EntityChangeEvent<SPSController>) => void;
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

export function toRefreshKeyResolver(
  refreshKey?: RefreshKeyInput
): () => string | number | undefined {
  if (typeof refreshKey === 'function') {
    return refreshKey;
  }

  return () => refreshKey;
}
