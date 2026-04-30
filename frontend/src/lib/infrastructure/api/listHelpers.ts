import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export interface ApiPaginatedListResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit?: number;
  total_pages: number;
}

export function buildListSearchParams(params: ListParams): URLSearchParams {
  const searchParams = new URLSearchParams();
  searchParams.set('page', String(params.pagination.page));
  searchParams.set('limit', String(params.pagination.pageSize));

  if (params.search.text) {
    searchParams.set('search', params.search.text);
  }

  if (params.filters) {
    for (const [key, value] of Object.entries(params.filters)) {
      if (value !== undefined && value !== null && value !== '') {
        searchParams.set(key, value);
      }
    }
  }

  return searchParams;
}

export function buildListUrl(path: string, params: ListParams): string {
  const query = buildListSearchParams(params).toString();
  return `${path}${query ? `?${query}` : ''}`;
}

export function mapPaginatedResponse<T>(
  response: ApiPaginatedListResponse<T>,
  params: ListParams
): PaginatedResponse<T> {
  return {
    items: response.items,
    metadata: {
      total: response.total,
      page: response.page,
      pageSize: response.limit ?? params.pagination.pageSize,
      totalPages: response.total_pages
    }
  };
}
