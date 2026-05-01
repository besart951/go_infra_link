import type { FieldDevice } from '$lib/domain/facility/index.js';
import { ContextualDataTableFetchStrategy } from '$lib/state/table/ContextualDataTableFetchStrategy.js';
import type { FieldDeviceFilters } from '../types.js';
import { FacilityFieldDeviceFetchStrategy } from './FacilityFieldDeviceFetchStrategy.js';
import { ProjectFieldDeviceFetchStrategy } from './ProjectFieldDeviceFetchStrategy.js';

export class ContextualFieldDeviceFetchStrategy extends ContextualDataTableFetchStrategy<
  FieldDevice,
  FieldDeviceFilters,
  ProjectFieldDeviceFetchStrategy
> {
  constructor(resolveProjectId: () => string | undefined) {
    super(
      resolveProjectId,
      new FacilityFieldDeviceFetchStrategy(),
      (projectId) => new ProjectFieldDeviceFetchStrategy(projectId)
    );
  }
}
