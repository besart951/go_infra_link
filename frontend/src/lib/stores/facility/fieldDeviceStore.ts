import { writable, get } from 'svelte/store';
import { api } from '$lib/api/client.js';
import type { FieldDevice } from '$lib/domain/facility/index.js';

/**
 * Field device filter parameters
 */
export interface FieldDeviceFilters {
	buildingId?: string;
	controlCabinetId?: string;
	spsControllerId?: string;
	spsControllerSystemTypeId?: string;
	projectId?: string;
}

/**
 * Field device list state
 */
export interface FieldDeviceListState {
	items: FieldDevice[];
	total: number;
	page: number;
	pageSize: number;
	totalPages: number;
	searchText: string;
	orderBy?: string;
	order?: 'asc' | 'desc';
	filters: FieldDeviceFilters;
	loading: boolean;
	error: string | null;
}

/**
 * Backend response format
 */
interface BackendListResponse {
	items: FieldDevice[];
	total: number;
	page: number;
	total_pages: number;
}

/**
 * Creates a field device store with filtering support
 * @param pageSize - Number of items per page
 * @param projectId - Optional project ID to scope field devices to a project
 */
export function createFieldDeviceStore(pageSize = 300, projectId?: string) {
	const initialState: FieldDeviceListState = {
		items: [],
		total: 0,
		page: 1,
		pageSize,
		totalPages: 0,
		searchText: '',
		orderBy: undefined,
		order: undefined,
		filters: {},
		loading: false,
		error: null
	};

	const store = writable<FieldDeviceListState>(initialState);
	let abortController: AbortController | null = null;

	/**
	 * Load field devices with current state
	 */
	async function load() {
		const state = get(store);

		// Cancel any ongoing request
		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		store.update((s) => ({ ...s, loading: true, error: null }));

		try {
			const searchParams = new URLSearchParams();
			searchParams.set('page', state.page.toString());
			searchParams.set('limit', state.pageSize.toString());

			if (state.searchText) {
				searchParams.set('search', state.searchText);
			}

			if (state.orderBy) {
				searchParams.set('order_by', state.orderBy);
			}
			if (state.order) {
				searchParams.set('order', state.order);
			}

			// Add filter parameters
			if (state.filters.buildingId) {
				searchParams.set('building_id', state.filters.buildingId);
			}
			if (state.filters.controlCabinetId) {
				searchParams.set('control_cabinet_id', state.filters.controlCabinetId);
			}
			if (state.filters.spsControllerId) {
				searchParams.set('sps_controller_id', state.filters.spsControllerId);
			}
			if (state.filters.spsControllerSystemTypeId) {
				searchParams.set('sps_controller_system_type_id', state.filters.spsControllerSystemTypeId);
			}

			const effectiveProjectId = projectId ?? state.filters.projectId;
			if (effectiveProjectId) {
				searchParams.set('project_id', effectiveProjectId);
			}

			const query = searchParams.toString();
			const url = `/facility/field-devices?${query}`;

			const response = await api<BackendListResponse>(url, { signal: abortController.signal });

			store.update((s) => ({
				...s,
				items: response.items,
				total: response.total,
				page: response.page,
				totalPages: response.total_pages,
				loading: false,
				error: null
			}));
		} catch (error: any) {
			if (error.name === 'AbortError') {
				return; // Request was cancelled
			}
			store.update((s) => ({
				...s,
				loading: false,
				error: error.message || 'Failed to load field devices'
			}));
		}
	}

	return {
		subscribe: store.subscribe,

		/**
		 * Load field devices
		 */
		load,

		/**
		 * Search field devices
		 */
		search: (text: string) => {
			store.update((s) => ({ ...s, searchText: text, page: 1 }));
			load();
		},

		/**
		 * Go to specific page
		 */
		goToPage: (page: number) => {
			store.update((s) => ({ ...s, page }));
			load();
		},

		/**
		 * Reload current page
		 */
		reload: () => {
			load();
		},

		/**
		 * Update filters and reload
		 */
		setFilters: (filters: FieldDeviceFilters) => {
			store.update((s) => ({ ...s, filters, page: 1 }));
			load();
		},

		/**
		 * Update sorting and reload
		 */
		setSort: (orderBy?: string, order?: 'asc' | 'desc') => {
			store.update((s) => ({ ...s, orderBy, order, page: 1 }));
			load();
		},

		/**
		 * Clear a specific filter
		 */
		clearFilter: (filterKey: keyof FieldDeviceFilters) => {
			store.update((s) => {
				const newFilters = { ...s.filters };
				delete newFilters[filterKey];
				return { ...s, filters: newFilters, page: 1 };
			});
			load();
		},

		/**
		 * Clear all filters
		 */
		clearAllFilters: () => {
			store.update((s) => ({ ...s, filters: {}, page: 1 }));
			load();
		},

		/**
		 * Reset to initial state
		 */
		reset: () => {
			store.set(initialState);
		},

		/**
		 * Replace a field device in the current page
		 */
		updateItem: (updated: FieldDevice) => {
			store.update((s) => ({
				...s,
				items: s.items.map((item) => (item.id === updated.id ? updated : item))
			}));
		}
	};
}

/**
 * Global field device store instance
 */
export const fieldDeviceStore = createFieldDeviceStore();
