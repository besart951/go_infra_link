/**
 * Domain types for Field Device Multi-Create
 * Following hexagonal architecture - pure domain logic, no framework dependencies
 */

/**
 * Represents a single field device row in the multi-create form
 */
export interface FieldDeviceRowData {
	id: string;
	bmk: string;
	description: string;
	apparatNr: number | null;
}

/**
 * Validation error for a field device row
 */
export interface FieldDeviceRowError {
	message: string;
	field: 'bmk' | 'description' | 'apparat_nr' | '';
}

/**
 * Selection state for the multi-create form
 */
export interface MultiCreateSelection {
	spsControllerSystemTypeId: string;
	objectDataId: string;
	apparatId: string;
	systemPartId: string;
}

/**
 * Creates a unique key from a selection for comparison
 */
export function createSelectionKey(selection: MultiCreateSelection): string {
	return `${selection.spsControllerSystemTypeId}|${selection.objectDataId}|${selection.apparatId}|${selection.systemPartId}`;
}

/**
 * Checks if all required selections are made
 */
export function hasRequiredSelections(selection: MultiCreateSelection): boolean {
	return Boolean(
		selection.spsControllerSystemTypeId &&
			selection.objectDataId &&
			selection.apparatId &&
			selection.systemPartId
	);
}

/**
 * Checks if the minimum selections for fetching available numbers are made
 */
export function canFetchAvailableNumbers(selection: MultiCreateSelection): boolean {
	return Boolean(
		selection.spsControllerSystemTypeId && selection.apparatId && selection.systemPartId
	);
}

/**
 * Creates a new empty row with the next available apparat number
 */
export function createNewRow(availableNumbers: number[], usedNumbers: Set<number>): FieldDeviceRowData | null {
	const nextAvailable = availableNumbers.find((nr) => !usedNumbers.has(nr));
	if (nextAvailable === undefined) {
		return null;
	}

	return {
		id: crypto.randomUUID(),
		bmk: '',
		description: '',
		apparatNr: nextAvailable
	};
}

/**
 * Gets all used apparat numbers from rows
 */
export function getUsedApparatNumbers(rows: FieldDeviceRowData[]): Set<number> {
	const used = new Set<number>();
	for (const row of rows) {
		if (row.apparatNr !== null) {
			used.add(row.apparatNr);
		}
	}
	return used;
}

/**
 * Validates a single row's apparat number
 * Returns error or null if valid
 */
export function validateApparatNr(
	apparatNr: number | null,
	rowIndex: number,
	availableNumbers: number[],
	allRows: FieldDeviceRowData[],
	requireValue: boolean
): FieldDeviceRowError | null {
	// Check if value is required
	if (apparatNr === null) {
		if (requireValue) {
			return { message: 'Apparat number is required', field: 'apparat_nr' };
		}
		return null;
	}

	// Check range
	if (apparatNr < 1 || apparatNr > 99) {
		return { message: 'Apparat number must be between 1 and 99', field: 'apparat_nr' };
	}

	// Check if available (not used by existing field devices in DB)
	if (!availableNumbers.includes(apparatNr)) {
		return { message: 'This apparat number is already used', field: 'apparat_nr' };
	}

	// Check for duplicates within the form (excluding current row)
	const duplicateIndex = allRows.findIndex(
		(r, i) => i !== rowIndex && r.apparatNr === apparatNr
	);
	if (duplicateIndex !== -1) {
		return {
			message: `Duplicate: also used in row #${duplicateIndex + 1}`,
			field: 'apparat_nr'
		};
	}

	return null;
}

/**
 * Validates all rows and returns a map of row index to error
 */
export function validateAllRows(
	rows: FieldDeviceRowData[],
	availableNumbers: number[],
	requireValues: boolean
): Map<number, FieldDeviceRowError> {
	const errors = new Map<number, FieldDeviceRowError>();
	
	for (let i = 0; i < rows.length; i++) {
		const error = validateApparatNr(rows[i].apparatNr, i, availableNumbers, rows, requireValues);
		if (error) {
			errors.set(i, error);
		}
	}
	
	return errors;
}

/**
 * Session storage key for persisting form state
 */
export const STORAGE_KEY = 'fieldDeviceMultiCreate';

/**
 * Persisted state structure
 */
export interface PersistedState {
	selection: MultiCreateSelection;
	rows: FieldDeviceRowData[];
}

/**
 * Loads persisted state from session storage
 */
export function loadPersistedState(): PersistedState | null {
	if (typeof sessionStorage === 'undefined') return null;
	
	const stored = sessionStorage.getItem(STORAGE_KEY);
	if (!stored) return null;
	
	try {
		const parsed = JSON.parse(stored);
		return {
			selection: {
				spsControllerSystemTypeId: parsed.spsControllerSystemTypeId || '',
				objectDataId: parsed.objectDataId || '',
				apparatId: parsed.apparatId || '',
				systemPartId: parsed.systemPartId || ''
			},
			rows: Array.isArray(parsed.rows)
				? parsed.rows.map((r: unknown) => {
						const row = r as Record<string, unknown>;
						return {
							id: typeof row?.id === 'string' ? row.id : crypto.randomUUID(),
							bmk: typeof row?.bmk === 'string' ? row.bmk : '',
							description: typeof row?.description === 'string' ? row.description : '',
							apparatNr: typeof row?.apparatNr === 'number' ? row.apparatNr : null
						};
					})
				: []
		};
	} catch {
		return null;
	}
}

/**
 * Saves state to session storage
 */
export function savePersistedState(selection: MultiCreateSelection, rows: FieldDeviceRowData[]): void {
	if (typeof sessionStorage === 'undefined') return;
	
	try {
		sessionStorage.setItem(
			STORAGE_KEY,
			JSON.stringify({
				spsControllerSystemTypeId: selection.spsControllerSystemTypeId,
				objectDataId: selection.objectDataId,
				apparatId: selection.apparatId,
				systemPartId: selection.systemPartId,
				rows
			})
		);
	} catch (err) {
		console.error('Failed to persist state:', err);
	}
}

/**
 * Clears persisted state
 */
export function clearPersistedState(): void {
	if (typeof sessionStorage === 'undefined') return;
	sessionStorage.removeItem(STORAGE_KEY);
}
