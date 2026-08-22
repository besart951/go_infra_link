import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import type { DataTableQuery } from '$lib/state/table/contracts.js';
import type { FieldDeviceCursorPage } from '$lib/domain/ports/facility/fieldDeviceRepository.js';
import type { FieldDeviceFilters } from './types.js';
import { buildFieldDeviceRequestFilters } from './buildFieldDeviceRequestFilters.js';

interface FieldDeviceCursorFetch {
  query: DataTableQuery<FieldDeviceFilters>;
  cursor?: string;
  projectId?: string;
  signal?: AbortSignal;
}

export function fetchFieldDeviceCursorPage(
  request: FieldDeviceCursorFetch
): Promise<FieldDeviceCursorPage> {
  const { query, cursor, projectId, signal } = request;
  return fieldDeviceRepository.listCursor(
    {
      limit: query.pageSize,
      ...(cursor ? { cursor } : {}),
      ...(query.searchText ? { search: query.searchText } : {}),
      ...(query.orderBy ? { orderBy: query.orderBy } : {}),
      ...(query.order ? { order: query.order } : {}),
      filters: buildFieldDeviceRequestFilters(query, projectId)
    },
    signal
  );
}
