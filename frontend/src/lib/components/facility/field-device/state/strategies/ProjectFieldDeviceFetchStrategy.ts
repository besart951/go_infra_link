import type { FieldDevice } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from '../types.js';
import { fetchFieldDevices } from '../fetchFieldDevices.js';

export class ProjectFieldDeviceFetchStrategy implements DataTableFetchStrategy<
  FieldDevice,
  FieldDeviceFilters
> {
  private readonly projectId: string;

  constructor(projectId: string) {
    this.projectId = projectId;
  }

  fetch(query: DataTableQuery<FieldDeviceFilters>, signal?: AbortSignal) {
    return fetchFieldDevices(query, signal, this.projectId);
  }
}
