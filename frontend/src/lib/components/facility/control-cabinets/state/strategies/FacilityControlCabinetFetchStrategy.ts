import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { ControlCabinetFilters } from '../types.js';

export class FacilityControlCabinetFetchStrategy implements DataTableFetchStrategy<
  ControlCabinet,
  ControlCabinetFilters
> {
  async fetch(query: DataTableQuery<ControlCabinetFilters>, signal?: AbortSignal) {
    const response = await controlCabinetRepository.list(
      {
        pagination: { page: query.page, pageSize: query.pageSize },
        search: { text: query.searchText }
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
}
