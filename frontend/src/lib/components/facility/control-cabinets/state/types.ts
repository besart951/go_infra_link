import type { TableFilterRecord } from '$lib/state/table/contracts.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { EntityChangeEvent } from '../../shared/entityRefresh.js';

export interface ControlCabinetFilters extends TableFilterRecord {}

export type ProjectIdInput = string | undefined | (() => string | undefined);

export interface ControlCabinetStateProps {
  projectId?: ProjectIdInput;
  pageSize?: number;
  onChanged?: (event?: EntityChangeEvent<ControlCabinet>) => void;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}
