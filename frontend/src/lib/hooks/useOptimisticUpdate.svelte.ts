/**
 * useOptimisticUpdate - Helper for implementing optimistic UI updates
 *
 * This composable provides utilities for optimistic updates:
 * - Immediately update UI before server response
 * - Automatically rollback on error
 * - Handle success/error callbacks
 */

export interface OptimisticUpdateOptions<T> {
	onSuccess?: (result: T) => void;
	onError?: (error: any) => void;
	onRollback?: () => void;
}

/**
 * Create an optimistic update handler
 */
export function useOptimisticUpdate<T>(options: OptimisticUpdateOptions<T> = {}) {
	let isOptimistic = $state(false);

	/**
	 * Execute an optimistic update
	 * @param optimisticAction - Function to update UI optimistically (called immediately)
	 * @param serverAction - Async function that performs the actual server request
	 * @param rollbackAction - Function to rollback the optimistic change on error
	 */
	async function execute(
		optimisticAction: () => void,
		serverAction: () => Promise<T>,
		rollbackAction: () => void
	): Promise<T | undefined> {
		// Apply optimistic update immediately
		isOptimistic = true;
		optimisticAction();

		try {
			// Execute server action
			const result = await serverAction();
			isOptimistic = false;
			options.onSuccess?.(result);
			return result;
		} catch (error) {
			// Rollback on error
			isOptimistic = false;
			rollbackAction();
			options.onRollback?.();
			options.onError?.(error);
			throw error;
		}
	}

	return {
		get isOptimistic() {
			return isOptimistic;
		},
		execute
	};
}
