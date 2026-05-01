import type { ControlCabinet } from '$lib/domain/facility/index.js';
import { ContextualDataTableFetchStrategy } from '$lib/state/table/ContextualDataTableFetchStrategy.js';
import type { ControlCabinetFilters } from '../types.js';
import { FacilityControlCabinetFetchStrategy } from './FacilityControlCabinetFetchStrategy.js';
import { ProjectControlCabinetFetchStrategy } from './ProjectControlCabinetFetchStrategy.js';

export class ContextualControlCabinetFetchStrategy extends ContextualDataTableFetchStrategy<
  ControlCabinet,
  ControlCabinetFilters,
  ProjectControlCabinetFetchStrategy
> {
  constructor(resolveProjectId: () => string | undefined) {
    super(
      resolveProjectId,
      new FacilityControlCabinetFetchStrategy(),
      (projectId) => new ProjectControlCabinetFetchStrategy(projectId)
    );
  }

  getBuildingLabels(): Map<string, string> {
    return this.getActiveProjectStrategy()?.getBuildingLabels() ?? new Map();
  }

  getLinkId(controlCabinetId: string): string | undefined {
    return this.getActiveProjectStrategy()?.getLinkId(controlCabinetId);
  }
}
