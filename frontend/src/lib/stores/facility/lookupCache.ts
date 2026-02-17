import { writable, get } from 'svelte/store';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import type { Apparat, SystemPart } from '$lib/domain/facility/index.js';

/**
 * Cache configuration
 */
const CACHE_TTL_MS = 30_000; // 30 seconds

/**
 * Generic cache entry
 */
interface CacheEntry<T> {
	items: T[];
	lastFetch: number;
	loading: boolean;
	searchQuery: string;
}

/**
 * Cache state for lookups
 */
interface LookupCacheState {
	apparats: CacheEntry<Apparat>;
	systemParts: CacheEntry<SystemPart>;
	apparatById: Map<string, Apparat>;
	systemPartById: Map<string, SystemPart>;
}

const initialState: LookupCacheState = {
	apparats: { items: [], lastFetch: 0, loading: false, searchQuery: '' },
	systemParts: { items: [], lastFetch: 0, loading: false, searchQuery: '' },
	apparatById: new Map(),
	systemPartById: new Map()
};

const store = writable<LookupCacheState>(initialState);

/**
 * Check if cache is stale
 */
function isCacheStale(entry: CacheEntry<unknown>, searchQuery: string): boolean {
	const now = Date.now();
	// Cache is stale if TTL expired or search query changed
	return now - entry.lastFetch > CACHE_TTL_MS || entry.searchQuery !== searchQuery;
}

/**
 * Apparats cache functions
 */
async function fetchApparats(search: string = ''): Promise<Apparat[]> {
	const state = get(store);

	// Return cached if not stale and same search query
	if (!isCacheStale(state.apparats, search) && state.apparats.items.length > 0) {
		return state.apparats.items;
	}

	// If already loading, wait and return current items
	if (state.apparats.loading) {
		return state.apparats.items;
	}

	// Set loading
	store.update((s) => ({
		...s,
		apparats: { ...s.apparats, loading: true }
	}));

	try {
		const res = await apparatRepository.list({
			pagination: { page: 1, pageSize: 100 },
			search: { text: search }
		});
		const items = res.items;

		// Update cache and store items by ID
		store.update((s) => {
			const newById = new Map(s.apparatById);
			items.forEach((item) => newById.set(item.id, item));
			return {
				...s,
				apparats: {
					items,
					lastFetch: Date.now(),
					loading: false,
					searchQuery: search
				},
				apparatById: newById
			};
		});

		return items;
	} catch (error) {
		console.error('Failed to fetch apparats:', error);
		store.update((s) => ({
			...s,
			apparats: { ...s.apparats, loading: false }
		}));
		return get(store).apparats.items;
	}
}

async function fetchApparatById(id: string): Promise<Apparat | null> {
	const state = get(store);

	// Check local cache first
	const cached = state.apparatById.get(id);
	if (cached) {
		return cached;
	}

	// Check in items array
	const fromItems = state.apparats.items.find((a) => a.id === id);
	if (fromItems) {
		store.update((s) => {
			const newById = new Map(s.apparatById);
			newById.set(id, fromItems);
			return { ...s, apparatById: newById };
		});
		return fromItems;
	}

	// Fetch from API
	try {
		const item = await apparatRepository.get(id);
		if (item) {
			store.update((s) => {
				const newById = new Map(s.apparatById);
				newById.set(id, item);
				return { ...s, apparatById: newById };
			});
		}
		return item;
	} catch (error) {
		console.error('Failed to fetch apparat by ID:', error);
		return null;
	}
}

/**
 * System Parts cache functions
 */
async function fetchSystemParts(search: string = ''): Promise<SystemPart[]> {
	const state = get(store);

	// Return cached if not stale and same search query
	if (!isCacheStale(state.systemParts, search) && state.systemParts.items.length > 0) {
		return state.systemParts.items;
	}

	// If already loading, wait and return current items
	if (state.systemParts.loading) {
		return state.systemParts.items;
	}

	// Set loading
	store.update((s) => ({
		...s,
		systemParts: { ...s.systemParts, loading: true }
	}));

	try {
		const res = await systemPartRepository.list({
			pagination: { page: 1, pageSize: 100 },
			search: { text: search }
		});
		const items = res.items;

		// Update cache and store items by ID
		store.update((s) => {
			const newById = new Map(s.systemPartById);
			items.forEach((item) => newById.set(item.id, item));
			return {
				...s,
				systemParts: {
					items,
					lastFetch: Date.now(),
					loading: false,
					searchQuery: search
				},
				systemPartById: newById
			};
		});

		return items;
	} catch (error) {
		console.error('Failed to fetch system parts:', error);
		store.update((s) => ({
			...s,
			systemParts: { ...s.systemParts, loading: false }
		}));
		return get(store).systemParts.items;
	}
}

async function fetchSystemPartById(id: string): Promise<SystemPart | null> {
	const state = get(store);

	// Check local cache first
	const cached = state.systemPartById.get(id);
	if (cached) {
		return cached;
	}

	// Check in items array
	const fromItems = state.systemParts.items.find((sp) => sp.id === id);
	if (fromItems) {
		store.update((s) => {
			const newById = new Map(s.systemPartById);
			newById.set(id, fromItems);
			return { ...s, systemPartById: newById };
		});
		return fromItems;
	}

	// Fetch from API
	try {
		const item = await systemPartRepository.get(id);
		if (item) {
			store.update((s) => {
				const newById = new Map(s.systemPartById);
				newById.set(id, item);
				return { ...s, systemPartById: newById };
			});
		}
		return item;
	} catch (error) {
		console.error('Failed to fetch system part by ID:', error);
		return null;
	}
}

/**
 * Invalidate cache - useful after create/update/delete operations
 */
function invalidateApparats() {
	store.update((s) => ({
		...s,
		apparats: { ...s.apparats, lastFetch: 0 }
	}));
}

function invalidateSystemParts() {
	store.update((s) => ({
		...s,
		systemParts: { ...s.systemParts, lastFetch: 0 }
	}));
}

function invalidateAll() {
	store.update((s) => ({
		...s,
		apparats: { ...s.apparats, lastFetch: 0 },
		systemParts: { ...s.systemParts, lastFetch: 0 }
	}));
}

/**
 * Preload both caches - call this on page mount for better UX
 */
async function preloadAll() {
	await Promise.all([fetchApparats(''), fetchSystemParts('')]);
}

export const lookupCache = {
	subscribe: store.subscribe,

	// Apparats
	fetchApparats,
	fetchApparatById,
	invalidateApparats,

	// System Parts
	fetchSystemParts,
	fetchSystemPartById,
	invalidateSystemParts,

	// All
	invalidateAll,
	preloadAll
};
