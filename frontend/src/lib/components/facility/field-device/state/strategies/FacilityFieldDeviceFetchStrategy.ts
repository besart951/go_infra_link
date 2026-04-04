import type { FieldDevice } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from '../types.js';
import { fetchFieldDevices } from '../fetchFieldDevices.js';

export class FacilityFieldDeviceFetchStrategy implements DataTableFetchStrategy<
  FieldDevice,
  FieldDeviceFilters
> {
  fetch(query: DataTableQuery<FieldDeviceFilters>, signal?: AbortSignal) {
    return fetchFieldDevices(query, signal);
  }
}
