/**
 * useFormState - Reusable form state management using Svelte 5 runes
 *
 * This composable manages common form state:
 * - Loading state
 * - General error messages
 * - Field-level errors
 * - Submit handler wrapper with error handling
 * - Automatic toast notifications for errors
 */

import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';

export interface FormState {
	loading: boolean;
	error: string;
	fieldErrors: Record<string, string>;
}

export interface UseFormStateOptions {
	onSuccess?: (result: any) => void;
	onError?: (error: any) => void;
	showErrorToast?: boolean; // Default: true - show toast on general errors
	showSuccessToast?: boolean; // Default: false - don't show toast on success
	successMessage?: string; // Custom success message
}

/**
 * Create reactive form state
 */
export function useFormState(options: UseFormStateOptions = {}) {
	const {
		onSuccess,
		onError,
		showErrorToast = true,
		showSuccessToast = false,
		successMessage
	} = options;

	const state = $state<FormState>({
		loading: false,
		error: '',
		fieldErrors: {}
	});

	/**
	 * Reset all errors
	 */
	function resetErrors() {
		state.error = '';
		state.fieldErrors = {};
	}

	/**
	 * Get error for a specific field
	 */
	function getFieldError(name: string, prefixes: string[] = []): string | undefined {
		// Try direct field name
		if (state.fieldErrors[name]) {
			return state.fieldErrors[name];
		}

		// Try with prefixes
		for (const prefix of prefixes) {
			const key = `${prefix}.${name}`;
			if (state.fieldErrors[key]) {
				return state.fieldErrors[key];
			}
		}

		return undefined;
	}

	/**
	 * Wrap an async form submission handler with error handling
	 */
	async function handleSubmit<T>(submitFn: () => Promise<T>): Promise<T | undefined> {
		resetErrors();
		state.loading = true;

		try {
			const result = await submitFn();

			// Show success toast if enabled
			if (showSuccessToast) {
				addToast(successMessage || 'Operation completed successfully', 'success');
			}

			onSuccess?.(result);
			return result;
		} catch (e) {
			console.error('Form submission error:', e);
			state.fieldErrors = getFieldErrors(e);
			state.error = Object.keys(state.fieldErrors).length ? '' : getErrorMessage(e);

			// Show error toast if enabled and there are no field-specific errors
			if (showErrorToast && state.error) {
				addToast(state.error, 'error');
			}

			onError?.(e);
			return undefined;
		} finally {
			state.loading = false;
		}
	}

	return {
		get loading() {
			return state.loading;
		},
		get error() {
			return state.error;
		},
		get fieldErrors() {
			return state.fieldErrors;
		},
		resetErrors,
		getFieldError,
		handleSubmit
	};
}
