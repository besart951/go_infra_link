import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';

/**
 * Generic use case for creating, updating, and deleting entities.
 * Replaces per-entity ManageXxxUseCase boilerplate.
 *
 * For entities that need extra methods (validate, getDeleteImpact, etc.),
 * extend this class in an entity-specific use case.
 */
export class ManageEntityUseCase<T, C, U> {
	constructor(protected repository: CrudRepository<T, C, U>) {}

	async create(data: C, signal?: AbortSignal): Promise<T> {
		return this.repository.create(data, signal);
	}

	async update(id: string, data: U, signal?: AbortSignal): Promise<T> {
		return this.repository.update(id, data, signal);
	}

	async delete(id: string, signal?: AbortSignal): Promise<void> {
		return this.repository.delete(id, signal);
	}

	async get(id: string, signal?: AbortSignal): Promise<T> {
		return this.repository.get(id, signal);
	}
}
