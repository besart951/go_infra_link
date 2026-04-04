import type { FieldDevice } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from '../types.js';
import { FacilityFieldDeviceFetchStrategy } from './FacilityFieldDeviceFetchStrategy.js';
import { ProjectFieldDeviceFetchStrategy } from './ProjectFieldDeviceFetchStrategy.js';

export class ContextualFieldDeviceFetchStrategy implements DataTableFetchStrategy<
  FieldDevice,
  FieldDeviceFilters
> {
  private readonly facilityStrategy = new FacilityFieldDeviceFetchStrategy();
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  fetch(query: DataTableQuery<FieldDeviceFilters>, signal?: AbortSignal) {
    const projectId = this.resolveProjectId();
    if (!projectId) {
      return this.facilityStrategy.fetch(query, signal);
    }

    return new ProjectFieldDeviceFetchStrategy(projectId).fetch(query, signal);
  }
}
