import { api } from '$lib/api/client.js';
import type {
	ListRepository,
	PaginatedResponse,
	ListParams
} from '$lib/domain/ports/listRepository.js';
import type { PaginationMetadata } from '$lib/domain/valueObjects/pagination.js';

/**
 * Backend response format
 */
interface BackendListResponse<T> {
	items: T[];
	total: number;
	page: number;
	total_pages: number;
}

/**
 * Generic API adapter for paginated lists
 * Implements the ListRepository port using the backend API
 */
export class ApiListAdapter<T> implements ListRepository<T> {
	constructor(
		private endpoint: string,
		private searchParam: string = 'search'
	) { }

	/**
	 * List items with pagination and search
	 */
	async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<T>> {
		const searchParams = new URLSearchParams();
		searchParams.set('page', params.pagination.page.toString());
		searchParams.set('limit', params.pagination.pageSize.toString());

		if (params.search.text) {
			searchParams.set(this.searchParam, params.search.text);
		}

		if (params.filters) {
			Object.entries(params.filters).forEach(([key, value]) => {
				if (value !== undefined && value !== null) searchParams.set(key, value);
			});
		}

		const query = searchParams.toString();
		const url = query ? `${this.endpoint}?${query}` : this.endpoint;

		const response = await api<BackendListResponse<T>>(url, { signal });

		const metadata: PaginationMetadata = {
			total: response.total,
			page: response.page,
			pageSize: params.pagination.pageSize,
			totalPages: response.total_pages
		};

		return {
			items: response.items,
			metadata
		};
	}

	/**
	 * Get a single item by ID
	 */
	async get(id: string, signal?: AbortSignal): Promise<T> {
		return api<T>(`${this.endpoint}/${id}`, { signal });
	}
}

/**
 * Factory function to create API adapters for different entities
 */
export function createApiAdapter<T>(endpoint: string, searchParam = 'search'): ListRepository<T> {
	return new ApiListAdapter<T>(endpoint, searchParam);
}
