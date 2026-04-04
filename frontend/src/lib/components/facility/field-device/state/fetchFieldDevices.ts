import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import type { FieldDevice } from '$lib/domain/facility/index.js';
import type { DataTablePage, DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceFilters } from './types.js';
import { buildFieldDeviceRequestFilters } from './buildFieldDeviceRequestFilters.js';

export async function fetchFieldDevices(
  query: DataTableQuery<FieldDeviceFilters>,
  signal?: AbortSignal,
  projectId?: string
): Promise<DataTablePage<FieldDevice>> {
  const response = await fieldDeviceRepository.list(
    {
      pagination: { page: query.page, pageSize: query.pageSize },
      search: { text: query.searchText },
      filters: buildFieldDeviceRequestFilters(query, projectId)
    },
    signal
  );

  return {
    items: response.items,
    total: response.metadata.total,
    page: response.metadata.page,
    totalPages: response.metadata.totalPages
  };
}
