import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { ControlCabinetFilters } from '../types.js';
import { FacilityControlCabinetFetchStrategy } from './FacilityControlCabinetFetchStrategy.js';
import { ProjectControlCabinetFetchStrategy } from './ProjectControlCabinetFetchStrategy.js';

export class ContextualControlCabinetFetchStrategy implements DataTableFetchStrategy<
  ControlCabinet,
  ControlCabinetFilters
> {
  private readonly facilityStrategy = new FacilityControlCabinetFetchStrategy();
  private projectStrategy: ProjectControlCabinetFetchStrategy | null = null;
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  async fetch(query: DataTableQuery<ControlCabinetFilters>, signal?: AbortSignal) {
    return this.getActiveStrategy().fetch(query, signal);
  }

  getBuildingLabels(): Map<string, string> {
    if (!this.projectStrategy) {
      return new Map();
    }

    return this.projectStrategy.getBuildingLabels();
  }

  getLinkId(controlCabinetId: string): string | undefined {
    if (!this.projectStrategy) {
      return undefined;
    }

    return this.projectStrategy.getLinkId(controlCabinetId);
  }

  private getActiveStrategy():
    | FacilityControlCabinetFetchStrategy
    | ProjectControlCabinetFetchStrategy {
    const projectId = this.resolveProjectId();
    if (!projectId) {
      return this.facilityStrategy;
    }

    if (!this.projectStrategy || this.projectStrategy.getProjectId() !== projectId) {
      this.projectStrategy = new ProjectControlCabinetFetchStrategy(projectId);
    }

    return this.projectStrategy;
  }
}
