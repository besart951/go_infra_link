import type { SPSController } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { SPSControllerFilters } from '../types.js';
import { FacilitySPSControllerFetchStrategy } from './FacilitySPSControllerFetchStrategy.js';
import { ProjectSPSControllerFetchStrategy } from './ProjectSPSControllerFetchStrategy.js';

export class ContextualSPSControllerFetchStrategy implements DataTableFetchStrategy<
  SPSController,
  SPSControllerFilters
> {
  private readonly facilityStrategy = new FacilitySPSControllerFetchStrategy();
  private projectStrategy: ProjectSPSControllerFetchStrategy | null = null;
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  async fetch(query: DataTableQuery<SPSControllerFilters>, signal?: AbortSignal) {
    return this.getActiveStrategy().fetch(query, signal);
  }

  getCabinetLabels(): Map<string, string> {
    if (!this.projectStrategy) {
      return new Map();
    }

    return this.projectStrategy.getCabinetLabels();
  }

  private getActiveStrategy():
    | FacilitySPSControllerFetchStrategy
    | ProjectSPSControllerFetchStrategy {
    const projectId = this.resolveProjectId();
    if (!projectId) {
      return this.facilityStrategy;
    }

    if (!this.projectStrategy || this.projectStrategy.getProjectId() !== projectId) {
      this.projectStrategy = new ProjectSPSControllerFetchStrategy(projectId);
    }

    return this.projectStrategy;
  }
}
