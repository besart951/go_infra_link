import type { ListRepository, ListParams, PaginatedResponse } from './listRepository.js';

/**
 * Generic CRUD repository port.
 * Extends ListRepository with standard get/create/update/delete operations.
 *
 * @typeParam T - Entity type
 * @typeParam C - Create request type
 * @typeParam U - Update request type
 */
export interface CrudRepository<T, C, U> extends ListRepository<T> {
	list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<T>>;
	get(id: string, signal?: AbortSignal): Promise<T>;
	getBulk?(ids: string[], signal?: AbortSignal): Promise<T[]>;
	create(data: C, signal?: AbortSignal): Promise<T>;
	update(id: string, data: U, signal?: AbortSignal): Promise<T>;
	delete(id: string, signal?: AbortSignal): Promise<void>;
}
