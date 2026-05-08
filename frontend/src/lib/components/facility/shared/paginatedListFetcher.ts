export interface LegacyPaginatedResponse<T> {
  items: T[];
  page?: number;
  total?: number;
  total_pages?: number;
  metadata?: {
    page?: number;
    totalPages?: number;
  };
}

export const DEFAULT_FACILITY_PAGE_SIZE = 1000;

export function resolveTotalPages<T>(
  response: LegacyPaginatedResponse<T>,
  pageSize: number
): number | undefined {
  if (typeof response.total_pages === 'number') return response.total_pages;
  if (typeof response.metadata?.totalPages === 'number') return response.metadata.totalPages;
  if (typeof response.total === 'number' && pageSize > 0) {
    return Math.ceil(response.total / pageSize);
  }

  return undefined;
}

export function resolvePage<T>(response: LegacyPaginatedResponse<T>): number | undefined {
  return response.page ?? response.metadata?.page;
}

export async function fetchAllPages<T>(
  fetchPage: (
    page: number,
    pageSize: number,
    signal?: AbortSignal
  ) => Promise<LegacyPaginatedResponse<T>>,
  signal?: AbortSignal,
  pageSize = DEFAULT_FACILITY_PAGE_SIZE
): Promise<T[]> {
  const allItems: T[] = [];
  let page = 1;

  while (true) {
    const response = await fetchPage(page, pageSize, signal);
    const items = response?.items ?? [];
    allItems.push(...items);

    const totalPages = resolveTotalPages(response, pageSize);
    if (typeof totalPages === 'number' && totalPages > 0) {
      if (page >= totalPages) break;
      page += 1;
      continue;
    }

    const responsePage = resolvePage(response) ?? page;
    if (items.length < pageSize) break;
    if (responsePage <= 0) break;

    page += 1;

    // Safety net for malformed or inconsistent APIs with constant page size responses.
    if (page > 10_000) {
      break;
    }
  }

  return allItems;
}
