/**
 * useFormState - Reusable form state management using Svelte 5 runes
 * 
 * This composable manages common form state:
 * - Loading state
 * - General error messages
 * - Field-level errors
 * - Submit handler wrapper with error handling
 */

import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';

export interface FormState {
	loading: boolean;
	error: string;
	fieldErrors: Record<string, string>;
}

export interface UseFormStateOptions {
	onSuccess?: (result: any) => void;
	onError?: (error: any) => void;
}

/**
 * Create reactive form state
 */
export function useFormState(options: UseFormStateOptions = {}) {
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
	async function handleSubmit<T>(
		submitFn: () => Promise<T>
	): Promise<T | undefined> {
		resetErrors();
		state.loading = true;

		try {
			const result = await submitFn();
			options.onSuccess?.(result);
			return result;
		} catch (e) {
			console.error('Form submission error:', e);
			state.fieldErrors = getFieldErrors(e);
			state.error = Object.keys(state.fieldErrors).length ? '' : getErrorMessage(e);
			options.onError?.(e);
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
