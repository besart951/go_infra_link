import type { SPSController } from '$lib/domain/facility/index.js';
import type { MultiFilterOption } from '$lib/components/facility/shared/projectFacilityListFilters.js';
import { ContextualDataTableFetchStrategy } from '$lib/state/table/ContextualDataTableFetchStrategy.js';
import type { SPSControllerFilters } from '../types.js';
import { FacilitySPSControllerFetchStrategy } from './FacilitySPSControllerFetchStrategy.js';
import { ProjectSPSControllerFetchStrategy } from './ProjectSPSControllerFetchStrategy.js';

export class ContextualSPSControllerFetchStrategy extends ContextualDataTableFetchStrategy<
  SPSController,
  SPSControllerFilters,
  ProjectSPSControllerFetchStrategy
> {
  constructor(resolveProjectId: () => string | undefined) {
    super(
      resolveProjectId,
      new FacilitySPSControllerFetchStrategy(),
      (projectId) => new ProjectSPSControllerFetchStrategy(projectId)
    );
  }

  getCabinetLabels(): Map<string, string> {
    return this.getActiveProjectStrategy()?.getCabinetLabels() ?? new Map();
  }

  getCabinetFilterOptions(): MultiFilterOption[] {
    return this.getActiveProjectStrategy()?.getCabinetFilterOptions() ?? [];
  }
}
