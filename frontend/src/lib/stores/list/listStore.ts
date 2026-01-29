import { writable, derived, get } from 'svelte/store';
import type { ListState } from '$lib/application/useCases/listUseCase.js';
import { ListUseCase } from '$lib/application/useCases/listUseCase.js';
import type { ListRepository } from '$lib/domain/ports/listRepository.js';
import { createPagination } from '$lib/domain/valueObjects/pagination.js';
import { createSearchQuery } from '$lib/domain/valueObjects/search.js';
import { ApiException } from '$lib/api/client.js';

/**
 * Cache entry with timestamp
 */
interface CacheEntry<T> {
	timestamp: number;
	data: ListState<T>;
}

/**
 * Options for creating a list store
 */
export interface ListStoreOptions {
	pageSize?: number;
	cacheTTL?: number; // Time-to-live for cache in milliseconds
	debounceMs?: number; // Debounce delay for search
}

/**
 * Creates a reactive Svelte store for managing paginated lists
 * Supports pagination, search, caching, and prevents duplicate requests
 */
export function createListStore<T>(repository: ListRepository<T>, options: ListStoreOptions = {}) {
	const { pageSize = 10, cacheTTL = 30000, debounceMs = 300 } = options;

	const useCase = new ListUseCase(repository);
	const initialState = useCase.createInitialState(pageSize);

	// Create the writable store
	const store = writable<ListState<T>>(initialState);

	// Cache for storing previous requests
	const cache = new Map<string, CacheEntry<T>>();

	// AbortController for canceling ongoing requests
	let abortController: AbortController | null = null;

	// Debounce timer for search
	let debounceTimer: ReturnType<typeof setTimeout> | null = null;

	/**
	 * Generate cache key from current state
	 */
	function getCacheKey(page: number, searchText: string): string {
		return JSON.stringify({ page, searchText, pageSize });
	}

	/**
	 * Load data from repository
	 */
	async function load(page: number, searchText: string, force = false) {
		const cacheKey = getCacheKey(page, searchText);

		// Check cache first
		if (!force && cacheTTL > 0) {
			const cached = cache.get(cacheKey);
			if (cached && Date.now() - cached.timestamp < cacheTTL) {
				store.set(cached.data);
				return;
			}
		}

		// Cancel any ongoing request
		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		// Set loading state
		store.update((s) => ({ ...s, loading: true, error: null }));

		try {
			const response = await useCase.execute(
				{
					pagination: createPagination(page, pageSize),
					search: createSearchQuery(searchText)
				},
				abortController.signal
			);

			const newState: ListState<T> = {
				items: response.items,
				total: response.metadata.total,
				page: response.metadata.page,
				pageSize: response.metadata.pageSize,
				totalPages: response.metadata.totalPages,
				searchText,
				loading: false,
				error: null
			};

			store.set(newState);

			// Cache the result
			cache.set(cacheKey, { timestamp: Date.now(), data: newState });
		} catch (error) {
			// Ignore abort errors
			if (error instanceof DOMException && error.name === 'AbortError') {
				return;
			}

			// authorization_failed is already surfaced via a toast in the central api client.
			// Keep the previous items and avoid showing an inline table error.
			if (error instanceof ApiException && error.error === 'authorization_failed') {
				store.update((s) => ({ ...s, loading: false, error: null }));
				return;
			}

			const errorMessage = error instanceof Error ? error.message : 'Failed to load data';
			store.update((s) => ({
				...s,
				loading: false,
				error: errorMessage
			}));
		} finally {
			abortController = null;
		}
	}

	/**
	 * Debounced load function for search
	 */
	function debouncedLoad(page: number, searchText: string) {
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		debounceTimer = setTimeout(() => {
			load(page, searchText);
		}, debounceMs);
	}

	return {
		subscribe: store.subscribe,

		/**
		 * Load the first page
		 */
		load: (searchText = '') => load(1, searchText),

		/**
		 * Reload current page (bypass cache)
		 */
		reload: () => {
			const state = get(store);
			load(state.page, state.searchText, true);
		},

		/**
		 * Go to a specific page
		 */
		goToPage: (page: number) => {
			const state = get(store);
			load(page, state.searchText);
		},

		/**
		 * Go to next page
		 */
		nextPage: () => {
			const state = get(store);
			if (state.page < state.totalPages) {
				load(state.page + 1, state.searchText);
			}
		},

		/**
		 * Go to previous page
		 */
		previousPage: () => {
			const state = get(store);
			if (state.page > 1) {
				load(state.page - 1, state.searchText);
			}
		},

		/**
		 * Update search text (debounced)
		 */
		search: (searchText: string) => {
			debouncedLoad(1, searchText);
		},

		/**
		 * Clear cache
		 */
		clearCache: () => {
			cache.clear();
		},

		/**
		 * Reset to initial state
		 */
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

/**
 * Derived store that extracts items from the list state
 */
export function getItems<T>(listStore: ReturnType<typeof createListStore<T>>) {
	return derived(listStore, ($state) => $state.items);
}

/**
 * Derived store that checks if list is loading
 */
export function isLoading<T>(listStore: ReturnType<typeof createListStore<T>>) {
	return derived(listStore, ($state) => $state.loading);
}
