import type {
	ListRepository,
	PaginatedResponse,
	ListParams
} from '$lib/domain/ports/listRepository.js';

/**
 * List state representing the current state of a list
 */
export interface ListState<T> {
	items: T[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	searchText: string;
	loading: boolean;
	error: string | null;
}

/**
 * Use case for listing and searching entities with pagination
 * This is framework-agnostic and can be used with any UI library
 */
export class ListUseCase<T> {
	constructor(private repository: ListRepository<T>) {}

	/**
	 * Execute the list operation
	 */
	async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<T>> {
		return this.repository.list(params, signal);
	}

	/**
	 * Get a single item by ID
	 */
	async getById(id: string, signal?: AbortSignal): Promise<T | null> {
		if (!this.repository.getById) {
			throw new Error('getById not implemented for this repository');
		}
		try {
			return await this.repository.getById(id, signal);
		} catch (error) {
			return null;
		}
	}

	/**
	 * Create initial empty state
	 */
	createInitialState(pageSize = 10): ListState<T> {
		return {
			items: [],
			total: 0,
			page: 1,
			pageSize,
			totalPages: 0,
			searchText: '',
			loading: false,
			error: null
		};
	}
}
