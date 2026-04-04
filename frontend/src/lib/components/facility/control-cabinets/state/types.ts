import type { TableFilterRecord } from '$lib/state/table/contracts.js';

export interface ControlCabinetFilters extends TableFilterRecord {}

export type ProjectIdInput = string | undefined | (() => string | undefined);

export interface ControlCabinetStateProps {
  projectId?: ProjectIdInput;
  pageSize?: number;
  onChanged?: () => void;
}

export function toProjectIdResolver(projectId?: ProjectIdInput): () => string | undefined {
  if (typeof projectId === 'function') {
    return projectId;
  }

  return () => projectId;
}
