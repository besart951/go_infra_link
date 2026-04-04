import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import type { SPSController } from '$lib/domain/facility/index.js';
import type { DataTableFetchStrategy, DataTableQuery } from '$lib/state/table/contracts.js';
import type { SPSControllerFilters } from '../types.js';

export class FacilitySPSControllerFetchStrategy implements DataTableFetchStrategy<
  SPSController,
  SPSControllerFilters
> {
  async fetch(query: DataTableQuery<SPSControllerFilters>, signal?: AbortSignal) {
    const response = await spsControllerRepository.list(
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
