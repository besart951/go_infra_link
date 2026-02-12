/**
 * SessionStorageService - Infrastructure layer for browser sessionStorage operations
 *
 * Provides type-safe, error-resilient session storage with automatic JSON serialization.
 * Follows hexagonal architecture: infrastructure concern, isolated from business logic.
 */

export interface SessionStorageAdapter {
	save<T>(key: string, value: T): void;
	load<T>(key: string): T | null;
	remove(key: string): void;
	clear(): void;
	has(key: string): boolean;
}

/**
 * Browser sessionStorage implementation
 * Handles serialization, errors, and SSR compatibility
 */
export class BrowserSessionStorage implements SessionStorageAdapter {
	private isAvailable(): boolean {
		return typeof sessionStorage !== 'undefined';
	}

	save<T>(key: string, value: T): void {
		if (!this.isAvailable()) {
			console.warn('[SessionStorage] Not available (SSR or disabled)');
			return;
		}

		try {
			const serialized = JSON.stringify(value);
			window.sessionStorage.setItem(key, serialized);
		} catch (error) {
			console.error(`[SessionStorage] Failed to save key "${key}":`, error);
		}
	}

	load<T>(key: string): T | null {
		if (!this.isAvailable()) return null;

		try {
			const serialized = window.sessionStorage.getItem(key);
			if (!serialized) return null;

			return JSON.parse(serialized) as T;
		} catch (error) {
			console.error(`[SessionStorage] Failed to load key "${key}":`, error);
			return null;
		}
	}

	remove(key: string): void {
		if (!this.isAvailable()) return;

		try {
			window.sessionStorage.removeItem(key);
		} catch (error) {
			console.error(`[SessionStorage] Failed to remove key "${key}":`, error);
		}
	}

	clear(): void {
		if (!this.isAvailable()) return;

		try {
			window.sessionStorage.clear();
		} catch (error) {
			console.error('[SessionStorage] Failed to clear:', error);
		}
	}

	has(key: string): boolean {
		if (!this.isAvailable()) return false;
		return window.sessionStorage.getItem(key) !== null;
	}
}

/**
 * Factory function for dependency injection
 */
export function createSessionStorage(): SessionStorageAdapter {
	return new BrowserSessionStorage();
}

/**
 * Default singleton instance for convenience
 */
export const sessionStorage = createSessionStorage();
