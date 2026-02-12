/**
 * useUnsavedChangesWarning - Browser navigation warning for unsaved changes
 *
 * Displays native browser confirmation dialog when user attempts to:
 * - Close the browser tab/window
 * - Refresh the page
 * - Navigate away from the page
 *
 * Clean separation of concerns: only handles browser-level warnings.
 * Usage: Call with a reactive boolean indicating if changes exist.
 */

export interface UnsavedChangesWarningOptions {
	/**
	 * Custom message for the warning dialog
	 * Note: Most modern browsers ignore custom messages and show a generic warning
	 */
	message?: string;

	/**
	 * Enable/disable the warning system
	 */
	enabled?: boolean;
}

/**
 * Hook to warn users about unsaved changes before leaving the page
 *
 * @param hasUnsavedChanges - Reactive boolean indicating if there are unsaved changes
 * @param options - Configuration options
 *
 * @example
 * ```ts
 * const editing = useFieldDeviceEditing();
 * useUnsavedChangesWarning(() => editing.hasUnsavedChanges);
 * ```
 */
export function useUnsavedChangesWarning(
	hasUnsavedChanges: () => boolean,
	options: UnsavedChangesWarningOptions = {}
) {
	const { message = 'You have unsaved changes. Are you sure you want to leave?', enabled = true } =
		options;

	if (!enabled) return;

	// Only run in browser environment
	if (typeof window === 'undefined') return;

	/**
	 * BeforeUnload event handler
	 * Triggered when user tries to leave the page
	 */
	function handleBeforeUnload(event: BeforeUnloadEvent): string | undefined {
		// Check if there are unsaved changes
		if (!hasUnsavedChanges()) return undefined;

		// Prevent default to show confirmation dialog
		event.preventDefault();

		// Set returnValue for older browsers
		event.returnValue = message;

		// Return message (modern browsers ignore this but it's required for some)
		return message;
	}

	// Register event listener
	$effect(() => {
		window.addEventListener('beforeunload', handleBeforeUnload);

		// Cleanup on component unmount
		return () => {
			window.removeEventListener('beforeunload', handleBeforeUnload);
		};
	});
}

/**
 * Utility to manually trigger browser leave confirmation
 * Useful for testing or custom navigation flows
 */
export function confirmLeaveWithUnsavedChanges(
	customMessage = 'You have unsaved changes. Do you want to leave without saving?'
): boolean {
	return confirm(customMessage);
}
