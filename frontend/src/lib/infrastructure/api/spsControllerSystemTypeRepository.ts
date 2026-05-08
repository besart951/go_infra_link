import type { SPSControllerSystemTypeRepository } from '$lib/domain/ports/facility/spsControllerSystemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  SPSControllerSystemType,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { ApiException, api } from '$lib/api/client.js';
import { clearCachedFetchById, createCachedFetchById } from './createCachedFetchById.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

interface ListCacheEntry {
  expiresAt: number;
  value?: PaginatedResponse<SPSControllerSystemType>;
  promise?: Promise<PaginatedResponse<SPSControllerSystemType>>;
}

const LIST_CACHE_TTL_MS = 60_000;
const listCache = new Map<string, ListCacheEntry>();

function normalizeFilterValue(value: string): string {
  const values = value
    .split('|')
    .map((item) => item.trim())
    .filter((item) => item.length > 0);

  if (values.length <= 1) {
    return values[0] ?? '';
  }

  return [...new Set(values)].sort().join('|');
}

function normalizeFilters(filters?: Record<string, string>): Record<string, string> | undefined {
  if (!filters) {
    return undefined;
  }

  const entries = Object.entries(filters)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .map(([key, value]) => [key, normalizeFilterValue(value)] as const)
    .sort(([left], [right]) => left.localeCompare(right));

  if (entries.length === 0) {
    return undefined;
  }

  return Object.fromEntries(entries);
}

function toNormalizedListParams(params: ListParams): ListParams {
  const filters = normalizeFilters(params.filters);
  return {
    pagination: {
      page: params.pagination.page,
      pageSize: params.pagination.pageSize
    },
    search: {
      text: params.search.text.trim()
    },
    ...(filters ? { filters } : {})
  };
}

function buildListCacheKey(params: ListParams): string {
  const searchParams = new URLSearchParams();
  searchParams.set('page', String(params.pagination.page));
  searchParams.set('limit', String(params.pagination.pageSize));

  if (params.search.text) {
    searchParams.set('search', params.search.text);
  }

  if (params.filters) {
    for (const [key, value] of Object.entries(params.filters)) {
      searchParams.set(key, value);
    }
  }

  return searchParams.toString();
}

function clearSystemTypeListCache(): void {
  listCache.clear();
}

async function listSystemTypesWithCache(
  params: ListParams,
  signal?: AbortSignal
): Promise<PaginatedResponse<SPSControllerSystemType>> {
  const normalized = toNormalizedListParams(params);
  const cacheKey = buildListCacheKey(normalized);
  const now = Date.now();
  const existing = listCache.get(cacheKey);

  if (existing && existing.expiresAt > now) {
    if (existing.promise) {
      return existing.promise;
    }
    return (
      existing.value || {
        items: [],
        metadata: {
          total: 0,
          page: normalized.pagination.page,
          pageSize: normalized.pagination.pageSize,
          totalPages: 0
        }
      }
    );
  }

  const promise = api<SPSControllerSystemTypeListResponse>(
    buildListUrl('/facility/sps-controller-system-types', normalized),
    { signal }
  ).then((response) => mapPaginatedResponse(response, normalized));

  listCache.set(cacheKey, {
    expiresAt: now + LIST_CACHE_TTL_MS,
    promise
  });

  try {
    const value = await promise;
    listCache.set(cacheKey, {
      expiresAt: Date.now() + LIST_CACHE_TTL_MS,
      value
    });
    return value;
  } catch (error) {
    listCache.delete(cacheKey);
    throw error;
  }
}

const getCached = createCachedFetchById('facility-sps-controller-system-types', (id: string) =>
  api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`)
);

export const spsControllerSystemTypeRepository: SPSControllerSystemTypeRepository = {
  async list(
    params: ListParams,
    signal?: AbortSignal
  ): Promise<PaginatedResponse<SPSControllerSystemType>> {
    return listSystemTypesWithCache(params, signal);
  },

  async get(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
    const response = await getCached(id);

    if (!response) {
      throw new ApiException(404, 'not_found', 'SPS controller system type not found');
    }

    return response;
  },

  async copy(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
    const result = await api<SPSControllerSystemType>(
      `/facility/sps-controller-system-types/${id}/copy`,
      {
        method: 'POST',
        signal
      }
    );
    clearSystemTypeListCache();
    clearCachedFetchById('facility-sps-controller-system-types');
    return result;
  },
  async delete(id: string, signal?: AbortSignal): Promise<void> {
    await api<void>(`/facility/sps-controller-system-types/${id}`, {
      method: 'DELETE',
      signal
    });
    clearSystemTypeListCache();
    clearCachedFetchById('facility-sps-controller-system-types');
  }
};
