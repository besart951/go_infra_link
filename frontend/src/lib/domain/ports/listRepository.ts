import type { Pagination, PaginationMetadata } from '../valueObjects/pagination.js';
import type { SearchQuery } from '../valueObjects/search.js';

/**
 * Generic paginated list response
 */
export interface PaginatedResponse<T> {
	items: T[];
	metadata: PaginationMetadata;
}

/**
 * List parameters combining pagination and search
 */
export interface ListParams {
	pagination: Pagination;
	search: SearchQuery;
}

/**
 * Repository port for fetching paginated lists
 * This is the interface that infrastructure adapters must implement
 */
export interface ListRepository<T> {
	/**
	 * List items with pagination and search
	 */
	list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<T>>;

	/**
	 * Get a single item by ID
	 */
	getById?(id: string, signal?: AbortSignal): Promise<T>;
}
