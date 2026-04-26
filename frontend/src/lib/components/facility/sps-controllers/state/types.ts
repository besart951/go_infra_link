import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import type { SPSController } from '$lib/domain/facility/index.js';
import type { EntityChangeEvent } from '../../shared/entityRefresh.js';

export interface SPSControllerFilters extends TableFilterRecord {}

export type ProjectIdInput = string | undefined | (() => string | undefined);
export type RefreshKeyInput = string | number | undefined | (() => string | number | undefined);

export interface SPSControllerStateProps {
  projectId?: ProjectIdInput;
  pageSize?: number;
  controlCabinetRefreshKey?: RefreshKeyInput;
  onChanged?: (event?: EntityChangeEvent<SPSController>) => void;
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
