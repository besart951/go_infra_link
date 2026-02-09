import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';

interface LiveValidationOptions {
	debounceMs?: number;
}

export function useLiveValidation<T>(
	validateFn: (payload: T) => Promise<void>,
	options: LiveValidationOptions = {}
) {
	const { debounceMs = 350 } = options;

	let fieldErrors = $state<Record<string, string>>({});
	let error = $state('');
	let validating = $state(false);
	let timer: ReturnType<typeof setTimeout> | null = null;

	function reset() {
		fieldErrors = {};
		error = '';
	}

	function trigger(payload: T) {
		if (timer) {
			clearTimeout(timer);
		}

		timer = setTimeout(async () => {
			validating = true;
			reset();
			try {
				await validateFn(payload);
			} catch (err) {
				fieldErrors = getFieldErrors(err);
				error = Object.keys(fieldErrors).length ? '' : getErrorMessage(err);
			} finally {
				validating = false;
			}
		}, debounceMs);
	}

	return {
		get fieldErrors() {
			return fieldErrors;
		},
		get error() {
			return error;
		},
		get validating() {
			return validating;
		},
		reset,
		trigger
	};
}
