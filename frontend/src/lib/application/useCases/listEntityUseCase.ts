import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

/**
 * Generic use case for listing, getting, and bulk-fetching entities.
 * Replaces per-entity ListXxxUseCase boilerplate.
 */
export class ListEntityUseCase<T, C = unknown, U = unknown> {
	constructor(private repository: CrudRepository<T, C, U>) {}

	async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<T>> {
		return this.repository.list(params, signal);
	}

	async get(id: string, signal?: AbortSignal): Promise<T> {
		return this.repository.get(id, signal);
	}

	async getBulk(ids: string[], signal?: AbortSignal): Promise<T[]> {
		if (!this.repository.getBulk) {
			throw new Error('getBulk not implemented for this repository');
		}
		return this.repository.getBulk(ids, signal);
	}
}
