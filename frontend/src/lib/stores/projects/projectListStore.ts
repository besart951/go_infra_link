import { writable, get } from 'svelte/store';
import type { ListState } from '$lib/application/useCases/listUseCase.js';
import type { PaginationMetadata } from '$lib/domain/valueObjects/pagination.js';
import type { Project, ProjectStatus, ProjectListParams } from '$lib/domain/project/index.js';
import { listProjects } from '$lib/infrastructure/api/project.adapter.js';

/**
 * Project status filter options
 */
export type ProjectStatusFilter = ProjectStatus | 'all';

/**
 * Project list state including status filter
 */
export type ProjectListState = ListState<Project> & {
	status: ProjectStatusFilter;
};

interface CacheEntry {
	timestamp: number;
	data: ProjectListState;
}

export interface ProjectListStoreOptions {
	pageSize?: number;
	cacheTTL?: number;
	debounceMs?: number;
}

export function createProjectListStore(options: ProjectListStoreOptions = {}) {
	const { pageSize = 10, cacheTTL = 30000, debounceMs = 300 } = options;

	const initialState: ProjectListState = {
		items: [],
		total: 0,
		page: 1,
		pageSize,
		totalPages: 0,
		searchText: '',
		loading: false,
		error: null,
		status: 'all'
	};

	const store = writable<ProjectListState>(initialState);
	const cache = new Map<string, CacheEntry>();
	let abortController: AbortController | null = null;
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	function getCacheKey(page: number, searchText: string, status: ProjectStatusFilter): string {
		return JSON.stringify({ page, searchText, status, pageSize });
	}

	function toParams(
		page: number,
		searchText: string,
		status: ProjectStatusFilter
	): ProjectListParams {
		return {
			page,
			limit: pageSize,
			search: searchText || undefined,
			status: status === 'all' ? undefined : status
		};
	}

	async function load(
		page: number,
		searchText: string,
		status: ProjectStatusFilter,
		force = false
	) {
		const cacheKey = getCacheKey(page, searchText, status);
		if (!force && cacheTTL > 0) {
			const cached = cache.get(cacheKey);
			if (cached && Date.now() - cached.timestamp < cacheTTL) {
				store.set(cached.data);
				return;
			}
		}

		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		store.update((s: ProjectListState) => ({ ...s, loading: true, error: null, status }));

		try {
			const response = await listProjects(toParams(page, searchText, status), {
				signal: abortController.signal
			});

			const metadata: PaginationMetadata = (response as { metadata?: PaginationMetadata })
				.metadata ?? {
				total: response.total ?? response.items?.length ?? 0,
				page: response.page ?? page,
				pageSize: response.limit ?? pageSize,
				totalPages:
					response.total_pages ??
					Math.ceil((response.total ?? response.items?.length ?? 0) / pageSize)
			};

			const newState: ProjectListState = {
				items: response.items ?? [],
				total: metadata.total,
				page: metadata.page,
				pageSize: metadata.pageSize,
				totalPages: metadata.totalPages,
				searchText,
				loading: false,
				error: null,
				status
			};

			store.set(newState);
			cache.set(cacheKey, { timestamp: Date.now(), data: newState });
		} catch (error) {
			if (error instanceof DOMException && error.name === 'AbortError') {
				return;
			}

			const errorMessage = error instanceof Error ? error.message : 'Failed to load projects';
			store.update((s: ProjectListState) => ({ ...s, loading: false, error: errorMessage, status }));
		} finally {
			abortController = null;
		}
	}

	function debouncedLoad(page: number, searchText: string, status: ProjectStatusFilter) {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}
		debounceTimer = setTimeout(() => load(page, searchText, status), debounceMs);
	}

	return {
		subscribe: store.subscribe,

		load: () => {
			const state = get(store);
			load(state.page, state.searchText, state.status);
		},

		reload: () => {
			const state = get(store);
			load(state.page, state.searchText, state.status, true);
		},

		goToPage: (page: number) => {
			const state = get(store);
			load(page, state.searchText, state.status);
		},

		nextPage: () => {
			const state = get(store);
			if (state.page < state.totalPages) {
				load(state.page + 1, state.searchText, state.status);
			}
		},

		previousPage: () => {
			const state = get(store);
			if (state.page > 1) {
				load(state.page - 1, state.searchText, state.status);
			}
		},

		search: (searchText: string) => {
			const state = get(store);
			debouncedLoad(1, searchText, state.status);
		},

		setStatus: (status: ProjectStatusFilter) => {
			const state = get(store);
			if (state.status === status) return;
			load(1, state.searchText, status, true);
		},

		clearCache: () => {
			cache.clear();
		},

		reset: () => {
			if (abortController) {
				abortController.abort();
			}
			if (debounceTimer) {
				clearTimeout(debounceTimer);
			}
			cache.clear();
			store.set(initialState);
		}
	};
}

export const projectListStore = createProjectListStore({ pageSize: 10 });
