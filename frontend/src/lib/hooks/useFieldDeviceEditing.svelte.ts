/**
 * useFieldDeviceEditing - Composable for field device inline editing state
 *
 * Extracts all editing state + mutation functions from FieldDeviceListView.
 * Pattern follows useFormState.svelte.ts (getter-based for Svelte 5 reactivity).
 */

import {
	BACNET_SOFTWARE_TYPES,
	BACNET_HARDWARE_TYPES
} from '$lib/domain/facility/bacnet-object.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { addToast } from '$lib/components/toast.svelte';
import { sessionStorage } from '$lib/services/sessionStorageService.js';
import type {
	FieldDevice,
	UpdateFieldDeviceRequest,
	BulkUpdateFieldDeviceItem,
	SpecificationInput,
	BacnetObjectInput,
	BacnetObjectPatchInput
} from '$lib/domain/facility/index.js';

export interface EditErrorInfo {
	message?: string;
	fields?: Record<string, string>;
}

/**
 * Persisted state structure for sessionStorage
 */
interface PersistedEditingState {
	edits: Array<[string, Partial<BulkUpdateFieldDeviceItem>]>;
	bacnetEdits: Array<[string, Array<[string, Partial<BacnetObjectInput>]>]>;
	timestamp: number;
}

const STORAGE_KEY_PREFIX = 'fielddevice-editing';

export function useFieldDeviceEditing(projectId?: string) {
	const storageKey = projectId ? `${STORAGE_KEY_PREFIX}-${projectId}` : STORAGE_KEY_PREFIX;

	// Load persisted state on initialization
	const persistedState = loadPersistedState(storageKey);

	// Pending edits state for inline editing
	let pendingEdits = $state<Map<string, Partial<BulkUpdateFieldDeviceItem>>>(
		persistedState ? new Map(persistedState.edits) : new Map()
	);

	// BACnet pending edits: deviceId -> (objectId -> partial edits)
	let pendingBacnetEdits = $state<Map<string, Map<string, Partial<BacnetObjectInput>>>>(
		persistedState
			? new Map(
					persistedState.bacnetEdits.map(([deviceId, entries]) => [deviceId, new Map(entries)])
				)
			: new Map()
	);

	// Error tracking state per device ID
	let editErrors = $state<Map<string, EditErrorInfo>>(new Map());

	// BACnet field errors from server: deviceId -> (objectId -> { field: error })
	let bacnetFieldErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());
	// BACnet client-side validation errors: deviceId -> (objectId -> { field: error })
	let bacnetClientErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());

	/**
	 * Persistence: Load state from sessionStorage on initialization
	 */
	function loadPersistedState(key: string): PersistedEditingState | null {
		const loaded = sessionStorage.load<PersistedEditingState>(key);
		if (!loaded) return null;

		// Check if state is stale (older than 24 hours)
		const MAX_AGE_MS = 24 * 60 * 60 * 1000;
		if (Date.now() - loaded.timestamp > MAX_AGE_MS) {
			sessionStorage.remove(key);
			return null;
		}

		return loaded;
	}

	/**
	 * Persistence: Save current state to sessionStorage
	 */
	function savePersistedState() {
		if (pendingEdits.size === 0 && pendingBacnetEdits.size === 0) {
			// No edits - clear storage
			sessionStorage.remove(storageKey);
			return;
		}

		const state: PersistedEditingState = {
			edits: Array.from(pendingEdits.entries()),
			bacnetEdits: Array.from(pendingBacnetEdits.entries()).map(([deviceId, objMap]) => [
				deviceId,
				Array.from(objMap.entries())
			]),
			timestamp: Date.now()
		};

		sessionStorage.save(storageKey, state);
	}

	/**
	 * Auto-save to sessionStorage whenever edits change
	 */
	$effect(() => {
		// Track pendingEdits and pendingBacnetEdits for changes
		const _editsSize = pendingEdits.size;
		const _bacnetSize = pendingBacnetEdits.size;

		// Save to sessionStorage
		savePersistedState();
	});

	function setEditError(deviceId: string, info?: EditErrorInfo) {
		const next = new Map(editErrors);
		if (info) {
			next.set(deviceId, info);
		} else {
			next.delete(deviceId);
		}
		editErrors = next;
	}

	function queueEdit(deviceId: string, field: keyof BulkUpdateFieldDeviceItem, value: unknown) {
		const existing = pendingEdits.get(deviceId) || {};
		pendingEdits = new Map(pendingEdits).set(deviceId, { ...existing, [field]: value });
		// Clear any existing error for this device when editing
		setEditError(deviceId);
	}

	function queueSpecEdit(deviceId: string, field: keyof SpecificationInput, value: unknown) {
		const existing = pendingEdits.get(deviceId) || {};
		const existingSpec = existing.specification || {};
		const newSpec = { ...existingSpec, [field]: value };
		pendingEdits = new Map(pendingEdits).set(deviceId, {
			...existing,
			specification: newSpec
		});
		// Clear any existing error
		setEditError(deviceId);
	}

	function buildSpecificationPatch(
		spec: SpecificationInput | undefined
	): SpecificationInput | undefined {
		if (!spec) return undefined;
		// Include all fields - preserve null values for deletion support
		// Only filter out empty strings, convert them to null for server-side deletion
		const patch: Record<string, unknown> = {};
		for (const [key, value] of Object.entries(spec)) {
			if (value === '') {
				// Empty string means user wants to delete this field
				patch[key] = null;
			} else if (value !== undefined) {
				patch[key] = value;
			}
		}
		return Object.keys(patch).length > 0 ? (patch as SpecificationInput) : undefined;
	}

	function buildUpdateForDevice(
		deviceId: string,
		storeItems: FieldDevice[],
		options: { includeBacnet: boolean }
	): BulkUpdateFieldDeviceItem | null {
		const changes = pendingEdits.get(deviceId);
		const bacnetEdits = pendingBacnetEdits.get(deviceId);
		const includeBacnet = options.includeBacnet && bacnetEdits && bacnetEdits.size > 0;
		const update: BulkUpdateFieldDeviceItem = { id: deviceId };
		let hasChanges = false;

		if (changes) {
			if ('bmk' in changes) {
				update.bmk = changes.bmk;
				hasChanges = true;
			}
			if ('description' in changes) {
				update.description = changes.description;
				hasChanges = true;
			}
			if ('apparat_nr' in changes) {
				update.apparat_nr = changes.apparat_nr;
				hasChanges = true;
			}
			if ('apparat_id' in changes) {
				update.apparat_id = changes.apparat_id;
				hasChanges = true;
			}
			if ('system_part_id' in changes) {
				update.system_part_id = changes.system_part_id;
				hasChanges = true;
			}

			const spec = buildSpecificationPatch(changes.specification);
			if (spec) {
				update.specification = spec;
				hasChanges = true;
			}
		}

		if (includeBacnet) {
			const device = storeItems.find((d) => d.id === deviceId);
			if (device) {
				update.bacnet_objects = buildBacnetObjectsPayload(device, bacnetEdits);
				hasChanges = true;
			}
		}

		return hasChanges ? update : null;
	}

	function applyEditsToDevice(
		device: FieldDevice,
		options: { includeBacnet: boolean }
	): FieldDevice {
		const changes = pendingEdits.get(device.id);
		const bacnetEdits = pendingBacnetEdits.get(device.id);
		let updated: FieldDevice = { ...device };

		if (changes) {
			if ('bmk' in changes) {
				updated = { ...updated, bmk: changes.bmk };
			}
			if ('description' in changes) {
				updated = { ...updated, description: changes.description };
			}
			if ('apparat_nr' in changes && changes.apparat_nr !== undefined) {
				updated = { ...updated, apparat_nr: String(changes.apparat_nr) };
			}
			if ('apparat_id' in changes) {
				updated = { ...updated, apparat_id: changes.apparat_id as string };
			}
			if ('system_part_id' in changes) {
				updated = { ...updated, system_part_id: changes.system_part_id as string };
			}

			const specPatch = buildSpecificationPatch(changes.specification);
			if (specPatch) {
				if (updated.specification) {
					// Update existing specification
					updated = {
						...updated,
						specification: { ...updated.specification, ...specPatch }
					};
				} else {
					// Create new specification optimistically
					updated = {
						...updated,
						specification: {
							id: '', // Temporary, will be filled on next refresh
							created_at: new Date().toISOString(),
							updated_at: new Date().toISOString(),
							field_device_id: updated.id,
							specification_supplier: undefined,
							specification_brand: undefined,
							specification_type: undefined,
							additional_info_motor_valve: undefined,
							additional_info_size: undefined,
							additional_information_installation_location: undefined,
							electrical_connection_ph: undefined,
							electrical_connection_acdc: undefined,
							electrical_connection_amperage: undefined,
							electrical_connection_power: undefined,
							electrical_connection_rotation: undefined,
							...specPatch
						}
					};
				}
			}
		}

		if (options.includeBacnet && bacnetEdits && bacnetEdits.size > 0 && device.bacnet_objects) {
			updated = {
				...updated,
				bacnet_objects: device.bacnet_objects.map((obj) => {
					const edits = bacnetEdits.get(obj.id);
					return edits ? { ...obj, ...edits } : obj;
				})
			};
		}

		return updated;
	}

	function validatePendingEdits(deviceId: string): EditErrorInfo | null {
		const changes = pendingEdits.get(deviceId);
		if (!changes) return null;

		const fields: Record<string, string> = {};
		const bmk = changes.bmk;
		if (bmk !== undefined && bmk !== null && String(bmk).length > 10) {
			fields['fielddevice.bmk'] = 'BMK must be 10 characters or less';
		}
		const description = changes.description;
		if (description !== undefined && description !== null && String(description).length > 250) {
			fields['fielddevice.description'] = 'Description must be 250 characters or less';
		}
		const apparatNr = changes.apparat_nr;
		if (apparatNr !== undefined && apparatNr !== null) {
			const nr = Number(apparatNr);
			if (Number.isNaN(nr) || nr < 1 || nr > 99) {
				fields['fielddevice.apparat_nr'] = 'Apparat Nr must be between 1 and 99';
			}
		}

		const spec = changes.specification;
		if (spec) {
			const checkMax = (key: keyof SpecificationInput, label: string) => {
				const value = spec[key];
				if (value !== undefined && value !== null && String(value).length > 250) {
					fields[`specification.${key}`] = `${label} must be 250 characters or less`;
				}
			};
			checkMax('specification_supplier', 'Supplier');
			checkMax('specification_brand', 'Brand');
			checkMax('specification_type', 'Type');
			checkMax('additional_info_motor_valve', 'Motor/Valve');
			checkMax('additional_information_installation_location', 'Install Location');

			const acdc = spec.electrical_connection_acdc;
			if (acdc !== undefined && acdc !== null && String(acdc).length !== 2) {
				fields['specification.electrical_connection_acdc'] = 'AC/DC must be 2 characters';
			}
		}

		if (Object.keys(fields).length > 0) {
			return { message: 'validation_error', fields };
		}
		return null;
	}

	function isFieldDirty(deviceId: string, field: keyof UpdateFieldDeviceRequest): boolean {
		const edit = pendingEdits.get(deviceId);
		return edit ? field in edit : false;
	}

	function isSpecFieldDirty(deviceId: string, field: keyof SpecificationInput): boolean {
		const edit = pendingEdits.get(deviceId);
		if (!edit) return false;
		const spec = edit.specification;
		return spec ? field in spec : false;
	}

	function getPendingValue(
		deviceId: string,
		field: keyof BulkUpdateFieldDeviceItem
	): string | undefined {
		const edit = pendingEdits.get(deviceId);
		if (!edit || !(field in edit)) return undefined;
		const val = edit[field];
		return val !== undefined ? String(val) : undefined;
	}

	function getPendingSpecValue(
		deviceId: string,
		field: keyof SpecificationInput
	): string | undefined {
		const edit = pendingEdits.get(deviceId);
		if (!edit) return undefined;
		const spec = edit.specification;
		if (!spec || !(field in spec)) return undefined;
		const val = spec[field];
		return val !== undefined ? String(val) : undefined;
	}

	function getFieldError(deviceId: string, field: string): string | undefined {
		const errorInfo = editErrors.get(deviceId);
		if (!errorInfo) return undefined;

		if (errorInfo.fields && Object.keys(errorInfo.fields).length > 0) {
			if (errorInfo.fields[field]) return errorInfo.fields[field];
			if (errorInfo.fields[`fielddevice.${field}`]) return errorInfo.fields[`fielddevice.${field}`];
			if (errorInfo.fields[`specification.${field}`])
				return errorInfo.fields[`specification.${field}`];
			return undefined;
		}

		return undefined;
	}

	// BACnet edit queuing
	function queueBacnetEdit(deviceId: string, objectId: string, field: string, value: unknown) {
		const deviceEdits = pendingBacnetEdits.get(deviceId) || new Map();
		const objectEdits = deviceEdits.get(objectId) || {};
		deviceEdits.set(objectId, { ...objectEdits, [field]: value } as Partial<BacnetObjectInput>);
		pendingBacnetEdits = new Map(pendingBacnetEdits).set(deviceId, new Map(deviceEdits));
		clearBacnetFieldError(deviceId, objectId, field);
	}

	function clearBacnetFieldError(deviceId: string, objectId: string, field: string) {
		const deviceServerErrs = bacnetFieldErrors.get(deviceId);
		if (deviceServerErrs) {
			const objErrs = deviceServerErrs.get(objectId);
			if (objErrs && field in objErrs) {
				const { [field]: _, ...rest } = objErrs;
				const newDeviceErrs = new Map(deviceServerErrs);
				if (Object.keys(rest).length > 0) {
					newDeviceErrs.set(objectId, rest);
				} else {
					newDeviceErrs.delete(objectId);
				}
				bacnetFieldErrors = new Map(bacnetFieldErrors).set(deviceId, newDeviceErrs);
			}
		}
		const deviceClientErrs = bacnetClientErrors.get(deviceId);
		if (deviceClientErrs) {
			const objErrs = deviceClientErrs.get(objectId);
			if (objErrs && field in objErrs) {
				const { [field]: _, ...rest } = objErrs;
				const newDeviceErrs = new Map(deviceClientErrs);
				if (Object.keys(rest).length > 0) {
					newDeviceErrs.set(objectId, rest);
				} else {
					newDeviceErrs.delete(objectId);
				}
				bacnetClientErrors = new Map(bacnetClientErrors).set(deviceId, newDeviceErrs);
			}
		}
	}

	function validateBacnetEdits(items: FieldDevice[], deviceId: string): boolean {
		const device = items.find((d) => d.id === deviceId);
		if (!device || !device.bacnet_objects) return true;

		const deviceEdits = pendingBacnetEdits.get(deviceId);
		if (!deviceEdits || deviceEdits.size === 0) return true;

		const errors = new Map<string, Record<string, string>>();
		const validSoftwareTypes = new Set<string>(BACNET_SOFTWARE_TYPES.map((t) => t.value));
		const validHardwareTypes = new Set<string>(BACNET_HARDWARE_TYPES.map((t) => t.value));

		// Collect effective text_fix values for uniqueness check
		const textFixMap = new Map<string, string>();
		for (const obj of device.bacnet_objects) {
			const edits = deviceEdits.get(obj.id);
			const effectiveTextFix =
				edits && 'text_fix' in edits ? (edits.text_fix as string) : obj.text_fix;
			if (effectiveTextFix) {
				const existing = textFixMap.get(effectiveTextFix);
				if (existing && existing !== obj.id) {
					const objErrors = errors.get(obj.id) || {};
					objErrors['text_fix'] = 'text_fix must be unique within the field device';
					errors.set(obj.id, objErrors);
				}
				textFixMap.set(effectiveTextFix, obj.id);
			}
		}

		// Validate individual edited fields
		for (const [objectId, edits] of deviceEdits) {
			const objErrors = errors.get(objectId) || {};

			if ('text_fix' in edits && !edits.text_fix) {
				objErrors['text_fix'] = 'text_fix is required';
			}
			if ('software_number' in edits) {
				const num = edits.software_number as number;
				if (num < 0 || num > 65535) {
					objErrors['software_number'] = 'Must be between 0 and 65535';
				}
			}
			if ('hardware_quantity' in edits) {
				const num = edits.hardware_quantity as number;
				if (num < 1 || num > 255) {
					objErrors['hardware_quantity'] = 'Must be between 1 and 255';
				}
			}
			if ('software_type' in edits && !validSoftwareTypes.has(edits.software_type as string)) {
				objErrors['software_type'] = 'Invalid software type';
			}
			if ('hardware_type' in edits && !validHardwareTypes.has(edits.hardware_type as string)) {
				objErrors['hardware_type'] = 'Invalid hardware type';
			}
			if ('text_individual' in edits) {
				const val = edits.text_individual as string | undefined;
				if (val && val.length > 250) {
					objErrors['text_individual'] = 'Must be 250 characters or less';
				}
			}

			if (Object.keys(objErrors).length > 0) {
				errors.set(objectId, objErrors);
			}
		}

		if (errors.size > 0) {
			bacnetClientErrors = new Map(bacnetClientErrors).set(deviceId, errors);
			return false;
		} else {
			const newClientErrors = new Map(bacnetClientErrors);
			newClientErrors.delete(deviceId);
			bacnetClientErrors = newClientErrors;
			return true;
		}
	}

	function buildBacnetObjectsPayload(
		device: FieldDevice,
		deviceEdits: Map<string, Partial<BacnetObjectInput>>
	): BacnetObjectPatchInput[] {
		if (!device.bacnet_objects) return [];
		const patches: BacnetObjectPatchInput[] = [];
		for (const [objectId, edits] of deviceEdits.entries()) {
			const patch: BacnetObjectPatchInput = { id: objectId };
			let hasChanges = false;

			if ('text_fix' in edits) {
				patch.text_fix = edits.text_fix as string;
				hasChanges = true;
			}
			if ('description' in edits) {
				patch.description = edits.description as string | undefined;
				hasChanges = true;
			}
			if ('gms_visible' in edits) {
				patch.gms_visible = edits.gms_visible as boolean;
				hasChanges = true;
			}
			if ('optional' in edits) {
				patch.optional = edits.optional as boolean;
				hasChanges = true;
			}
			if ('text_individual' in edits) {
				patch.text_individual = edits.text_individual as string | undefined;
				hasChanges = true;
			}
			if ('software_type' in edits) {
				patch.software_type = edits.software_type as string;
				hasChanges = true;
			}
			if ('software_number' in edits) {
				patch.software_number = edits.software_number as number;
				hasChanges = true;
			}
			if ('hardware_type' in edits) {
				patch.hardware_type = edits.hardware_type as string;
				hasChanges = true;
			}
			if ('hardware_quantity' in edits) {
				patch.hardware_quantity = edits.hardware_quantity as number;
				hasChanges = true;
			}
			if ('software_reference_id' in edits) {
				patch.software_reference_id = edits.software_reference_id as string | undefined;
				hasChanges = true;
			}
			if ('state_text_id' in edits) {
				patch.state_text_id = edits.state_text_id as string | undefined;
				hasChanges = true;
			}
			if ('notification_class_id' in edits) {
				patch.notification_class_id = edits.notification_class_id as string | undefined;
				hasChanges = true;
			}
			if ('alarm_definition_id' in edits) {
				patch.alarm_definition_id = edits.alarm_definition_id as string | undefined;
				hasChanges = true;
			}

			if (hasChanges) {
				patches.push(patch);
			}
		}

		return patches;
	}

	async function saveAllPendingEdits(
		storeItems: FieldDevice[],
		onSuccess?: (updated: FieldDevice[]) => void
	): Promise<void> {
		if (pendingEdits.size === 0 && pendingBacnetEdits.size === 0) return;

		// Run client-side validation for all BACnet edits first
		let hasClientErrors = false;
		for (const deviceId of pendingBacnetEdits.keys()) {
			if (!validateBacnetEdits(storeItems, deviceId)) {
				hasClientErrors = true;
			}
		}
		if (hasClientErrors) {
			addToast('Fix validation errors before saving.', 'error');
			return;
		}

		// Collect all device IDs that need updates
		const allDeviceIds = new Set([...pendingEdits.keys(), ...pendingBacnetEdits.keys()]);
		const updates: BulkUpdateFieldDeviceItem[] = [];
		const nextErrors = new Map(editErrors);

		for (const id of allDeviceIds) {
			const clientError = validatePendingEdits(id);
			if (clientError) {
				nextErrors.set(id, clientError);
				continue;
			}
			nextErrors.delete(id);

			const update = buildUpdateForDevice(id, storeItems, { includeBacnet: true });
			if (update) {
				updates.push(update);
			}
		}

		if (updates.length === 0) {
			editErrors = nextErrors;
			addToast('Fix validation errors before saving.', 'error');
			return;
		}

		const pendingSnapshot = new Map(pendingEdits);
		const pendingBacnetSnapshot = new Map(pendingBacnetEdits);

		try {
			const result = await fieldDeviceRepository.bulkUpdate({ updates });

			// Process results and track errors
			const newErrors = new Map(nextErrors);
			const newBacnetFieldErrors = new Map<string, Map<string, Record<string, string>>>();
			const successIds = new Set<string>();
			const partialSuccessIds = new Set<string>(); // Devices with some fields that succeeded

			for (const r of result.results) {
				if (r.success) {
					successIds.add(r.id);
				} else {
					if (r.error) {
						newErrors.set(r.id, { message: r.error, fields: r.fields });
					}
					// Parse field-level errors for BACnet objects
					if (r.fields) {
						const objErrors = new Map<string, Record<string, string>>();
						for (const [fieldPath, msg] of Object.entries(r.fields)) {
							const match = fieldPath.match(/^bacnet_objects\.([0-9a-f-]+)\.(.+)$/i);
							if (match) {
								const objId = match[1];
								const field = match[2];
								const existing = objErrors.get(objId) || {};
								existing[field] = msg;
								objErrors.set(objId, existing);
							}
						}
						if (objErrors.size > 0) {
							newBacnetFieldErrors.set(r.id, objErrors);
						}
						// Mark this device for partial success handling
						partialSuccessIds.add(r.id);
					}
				}
			}

			const optimisticUpdates: FieldDevice[] = [];
			for (const id of successIds) {
				const device = storeItems.find((item) => item.id === id);
				if (device) {
					optimisticUpdates.push(applyEditsToDevice(device, { includeBacnet: true }));
				}
			}

			// Handle partial successes: apply non-failed fields optimistically
			for (const id of partialSuccessIds) {
				const device = storeItems.find((item) => item.id === id);
				if (!device) continue;

				const resultItem = result.results.find((r) => r.id === id);
				const failedFields = new Set<string>();
				const failedSpecFields = new Set<string>();
				const failedBacnetObjects = new Set<string>();
				let entireSpecificationFailed = false; // Generic spec error means all spec fields failed
				let entireFieldDeviceFailed = false; // Generic fielddevice error means all fields failed

				// Identify which fields failed
				if (resultItem?.fields) {
					for (const fieldPath of Object.keys(resultItem.fields)) {
						if (fieldPath === 'fielddevice') {
							// Generic fielddevice error - all base fields failed
							entireFieldDeviceFailed = true;
						} else if (fieldPath.startsWith('fielddevice.')) {
							failedFields.add(fieldPath.replace('fielddevice.', ''));
						} else if (fieldPath === 'specification') {
							// Generic specification error - all spec fields failed
							entireSpecificationFailed = true;
						} else if (fieldPath.startsWith('specification.')) {
							failedSpecFields.add(fieldPath.replace('specification.', ''));
						} else if (fieldPath.startsWith('bacnet_objects.')) {
							const match = fieldPath.match(/^bacnet_objects\.([0-9a-f-]+)/);
							if (match) failedBacnetObjects.add(match[1]);
						}
					}
				}

				// Build optimistic update with only successful fields
				const changes = pendingEdits.get(id);
				if (changes) {
					let updated: FieldDevice = { ...device };

					// Apply successful top-level fields (only if not all fielddevice fields failed)
					if (!entireFieldDeviceFailed) {
						if ('bmk' in changes && !failedFields.has('bmk')) {
							updated = { ...updated, bmk: changes.bmk };
						}
						if ('description' in changes && !failedFields.has('description')) {
							updated = { ...updated, description: changes.description };
						}
						if (
							'apparat_nr' in changes &&
							!failedFields.has('apparat_nr') &&
							changes.apparat_nr !== undefined
						) {
							updated = { ...updated, apparat_nr: String(changes.apparat_nr) };
						}
						if ('apparat_id' in changes && !failedFields.has('apparat_id')) {
							updated = { ...updated, apparat_id: changes.apparat_id as string };
						}
						if ('system_part_id' in changes && !failedFields.has('system_part_id')) {
							updated = { ...updated, system_part_id: changes.system_part_id as string };
						}
					}

					// Apply successful specification fields (only if not all spec fields failed)
					const specChanges = changes.specification;
					if (
						specChanges &&
						!entireSpecificationFailed &&
						failedSpecFields.size < Object.keys(specChanges).length
					) {
						const successfulSpecPatch: Record<string, unknown> = {};
						for (const [key, value] of Object.entries(specChanges)) {
							if (!failedSpecFields.has(key) && value !== undefined) {
								successfulSpecPatch[key] = value;
							}
						}
						if (Object.keys(successfulSpecPatch).length > 0) {
							if (updated.specification) {
								// Update existing specification
								updated = {
									...updated,
									specification: { ...updated.specification, ...successfulSpecPatch }
								};
							} else {
								// Create new specification (was created on backend, apply optimistically)
								updated = {
									...updated,
									specification: {
										id: '', // Will be populated on next full refresh
										created_at: new Date().toISOString(),
										updated_at: new Date().toISOString(),
										field_device_id: updated.id,
										specification_supplier: undefined,
										specification_brand: undefined,
										specification_type: undefined,
										additional_info_motor_valve: undefined,
										additional_info_size: undefined,
										additional_information_installation_location: undefined,
										electrical_connection_ph: undefined,
										electrical_connection_acdc: undefined,
										electrical_connection_amperage: undefined,
										electrical_connection_power: undefined,
										electrical_connection_rotation: undefined,
										...successfulSpecPatch
									}
								};
							}
						}
					}

					if (updated !== device) {
						optimisticUpdates.push(updated);
					}
				}

				// Apply successful BACnet object changes
				const bacnetEdits = pendingBacnetEdits.get(id);
				if (bacnetEdits && device.bacnet_objects) {
					let updated: FieldDevice = optimisticUpdates.find((u) => u.id === id) || { ...device };
					const updatedBacnetObjects = device.bacnet_objects.map((obj) => {
						if (failedBacnetObjects.has(obj.id)) return obj;
						const edits = bacnetEdits.get(obj.id);
						return edits ? { ...obj, ...edits } : obj;
					});
					updated = { ...updated, bacnet_objects: updatedBacnetObjects };
					const existingIndex = optimisticUpdates.findIndex((u) => u.id === id);
					if (existingIndex >= 0) {
						optimisticUpdates[existingIndex] = updated;
					} else {
						optimisticUpdates.push(updated);
					}
				}
			}

			// Remove successful edits from pending, keep failed ones
			const remainingEdits = new Map(pendingEdits);
			const remainingBacnetEdits = new Map(pendingBacnetEdits);

			// For fully successful devices, remove all pending edits
			for (const id of successIds) {
				if (pendingEdits.get(id) === pendingSnapshot.get(id)) {
					remainingEdits.delete(id);
				}
				if (pendingBacnetEdits.get(id) === pendingBacnetSnapshot.get(id)) {
					remainingBacnetEdits.delete(id);
				}
				newErrors.delete(id);
			}

			// For partially successful devices, remove only successful fields from pending
			for (const id of partialSuccessIds) {
				const resultItem = result.results.find((r) => r.id === id);
				if (!resultItem?.fields) continue;

				const failedFields = new Set<string>();
				const failedSpecFields = new Set<string>();
				let entireSpecificationFailed = false;
				let entireFieldDeviceFailed = false;

				for (const fieldPath of Object.keys(resultItem.fields)) {
					if (fieldPath === 'fielddevice') {
						entireFieldDeviceFailed = true;
					} else if (fieldPath.startsWith('fielddevice.')) {
						failedFields.add(fieldPath.replace('fielddevice.', ''));
					} else if (fieldPath === 'specification') {
						entireSpecificationFailed = true;
					} else if (fieldPath.startsWith('specification.')) {
						failedSpecFields.add(fieldPath.replace('specification.', ''));
					}
				}

				// Keep only failed fields in pending edits
				const changes = pendingEdits.get(id);
				if (changes) {
					const onlyFailedFields: Partial<BulkUpdateFieldDeviceItem> = {};

					// If entire fielddevice failed, keep all base fields
					if (entireFieldDeviceFailed) {
						if ('bmk' in changes) onlyFailedFields.bmk = changes.bmk;
						if ('description' in changes) onlyFailedFields.description = changes.description;
						if ('apparat_nr' in changes) onlyFailedFields.apparat_nr = changes.apparat_nr;
						if ('apparat_id' in changes) onlyFailedFields.apparat_id = changes.apparat_id;
						if ('system_part_id' in changes)
							onlyFailedFields.system_part_id = changes.system_part_id;
					} else {
						// Keep only specifically failed fields
						if ('bmk' in changes && failedFields.has('bmk')) {
							onlyFailedFields.bmk = changes.bmk;
						}
						if ('description' in changes && failedFields.has('description')) {
							onlyFailedFields.description = changes.description;
						}
						if ('apparat_nr' in changes && failedFields.has('apparat_nr')) {
							onlyFailedFields.apparat_nr = changes.apparat_nr;
						}
						if ('apparat_id' in changes && failedFields.has('apparat_id')) {
							onlyFailedFields.apparat_id = changes.apparat_id;
						}
						if ('system_part_id' in changes && failedFields.has('system_part_id')) {
							onlyFailedFields.system_part_id = changes.system_part_id;
						}
					}

					const specChanges = changes.specification;
					if (specChanges) {
						// If entire specification failed, keep all spec fields
						if (entireSpecificationFailed) {
							onlyFailedFields.specification = specChanges;
						} else if (failedSpecFields.size > 0) {
							// Keep only specifically failed spec fields
							const onlyFailedSpecFields: SpecificationInput = {};
							for (const [key, value] of Object.entries(specChanges)) {
								if (failedSpecFields.has(key)) {
									(onlyFailedSpecFields as Record<string, unknown>)[key] = value;
								}
							}
							if (Object.keys(onlyFailedSpecFields).length > 0) {
								onlyFailedFields.specification = onlyFailedSpecFields;
							}
						}
					}

					if (Object.keys(onlyFailedFields).length > 0) {
						remainingEdits.set(id, onlyFailedFields);
					} else {
						remainingEdits.delete(id);
					}
				}

				// For BACnet objects, keep only failed object edits
				const bacnetEdits = pendingBacnetEdits.get(id);
				if (bacnetEdits) {
					const failedBacnetObjects = new Set<string>();
					for (const fieldPath of Object.keys(resultItem.fields)) {
						const match = fieldPath.match(/^bacnet_objects\.([0-9a-f-]+)/);
						if (match) failedBacnetObjects.add(match[1]);
					}

					const onlyFailedBacnetEdits = new Map<string, Partial<BacnetObjectInput>>();
					for (const [objId, edits] of bacnetEdits.entries()) {
						if (failedBacnetObjects.has(objId)) {
							onlyFailedBacnetEdits.set(objId, edits);
						}
					}

					if (onlyFailedBacnetEdits.size > 0) {
						remainingBacnetEdits.set(id, onlyFailedBacnetEdits);
					} else {
						remainingBacnetEdits.delete(id);
					}
				}
			}

			pendingEdits = remainingEdits;
			pendingBacnetEdits = remainingBacnetEdits;
			editErrors = newErrors;
			bacnetFieldErrors = newBacnetFieldErrors;

			const totalSuccessful = successIds.size + partialSuccessIds.size;
			if (totalSuccessful > 0) {
				if (partialSuccessIds.size > 0) {
					addToast(
						`Updated ${successIds.size} device(s) completely, ${partialSuccessIds.size} device(s) partially. Check errors for failed fields.`,
						'warning'
					);
				} else {
					addToast(`Updated ${result.success_count} field device(s)`, 'success');
				}
				onSuccess?.(optimisticUpdates);
			}
			if (result.failure_count > 0 && partialSuccessIds.size === 0) {
				addToast(
					`Failed to update ${result.failure_count} device(s). Check highlighted fields.`,
					'error'
				);
			}
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Bulk update failed: ${err.message}`, 'error');
		}
	}

	async function saveDeviceEdits(
		device: FieldDevice,
		onSuccess?: (updated: FieldDevice) => void
	): Promise<void> {
		const update = buildUpdateForDevice(device.id, [device], { includeBacnet: false });
		if (!update) return;

		const clientError = validatePendingEdits(device.id);
		if (clientError) {
			setEditError(device.id, clientError);
			addToast('Fix validation errors before saving.', 'error');
			return;
		}
		setEditError(device.id);

		const pendingSnapshot = pendingEdits.get(device.id);

		const optimistic = applyEditsToDevice(device, { includeBacnet: false });

		try {
			const result = await fieldDeviceRepository.bulkUpdate({ updates: [update] });
			const item = result.results.find((r) => r.id === device.id);
			if (item?.success) {
				if (pendingEdits.get(device.id) === pendingSnapshot) {
					const remaining = new Map(pendingEdits);
					remaining.delete(device.id);
					pendingEdits = remaining;
				}
				setEditError(device.id);
				onSuccess?.(optimistic);
				return;
			}

			setEditError(device.id, { message: item?.error, fields: item?.fields });
			addToast(item?.error || 'Update failed. Check highlighted fields.', 'error');
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Update failed: ${err.message}`, 'error');
		}
	}

	async function saveDeviceBacnetEdits(
		device: FieldDevice,
		onSuccess?: (updated: FieldDevice) => void
	): Promise<void> {
		const update = buildUpdateForDevice(device.id, [device], { includeBacnet: true });
		if (!update) return;

		if (!validateBacnetEdits([device], device.id)) {
			addToast('Fix validation errors before saving.', 'error');
			return;
		}

		const clientError = validatePendingEdits(device.id);
		if (clientError) {
			setEditError(device.id, clientError);
			addToast('Fix validation errors before saving.', 'error');
			return;
		}
		setEditError(device.id);

		const pendingEditsSnapshot = pendingEdits.get(device.id);
		const pendingBacnetSnapshot = pendingBacnetEdits.get(device.id);
		const optimistic = applyEditsToDevice(device, { includeBacnet: true });

		try {
			const result = await fieldDeviceRepository.bulkUpdate({ updates: [update] });
			const item = result.results.find((r) => r.id === device.id);
			if (item?.success) {
				if (pendingEdits.get(device.id) === pendingEditsSnapshot) {
					const remaining = new Map(pendingEdits);
					remaining.delete(device.id);
					pendingEdits = remaining;
				}
				if (pendingBacnetEdits.get(device.id) === pendingBacnetSnapshot) {
					const remainingBacnet = new Map(pendingBacnetEdits);
					remainingBacnet.delete(device.id);
					pendingBacnetEdits = remainingBacnet;
				}

				const nextBacnetErrors = new Map(bacnetFieldErrors);
				nextBacnetErrors.delete(device.id);
				bacnetFieldErrors = nextBacnetErrors;
				onSuccess?.(optimistic);
				return;
			}

			const errorFields = item?.fields ?? {};
			const nextErrors = new Map(editErrors);
			nextErrors.set(device.id, { message: item?.error, fields: errorFields });
			editErrors = nextErrors;

			const objErrors = new Map<string, Record<string, string>>();
			for (const [fieldPath, msg] of Object.entries(errorFields)) {
				const match = fieldPath.match(/^bacnet_objects\.([0-9a-f-]+)\.(.+)$/i);
				if (!match) continue;
				const objId = match[1];
				const field = match[2];
				const existing = objErrors.get(objId) || {};
				existing[field] = msg;
				objErrors.set(objId, existing);
			}
			if (objErrors.size > 0) {
				const nextBacnetErrors = new Map(bacnetFieldErrors);
				nextBacnetErrors.set(device.id, objErrors);
				bacnetFieldErrors = nextBacnetErrors;
			}

			addToast(item?.error || 'Update failed. Check highlighted fields.', 'error');
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Update failed: ${err.message}`, 'error');
		}
	}

	function discardAllEdits() {
		pendingEdits = new Map();
		pendingBacnetEdits = new Map();
		editErrors = new Map();
		bacnetFieldErrors = new Map();
		bacnetClientErrors = new Map();
		// Clear persisted state from sessionStorage
		sessionStorage.remove(storageKey);
	}

	function getBacnetPendingEdits(deviceId: string): Map<string, Partial<BacnetObjectInput>> {
		return pendingBacnetEdits.get(deviceId) ?? new Map();
	}

	function getBacnetFieldErrors(deviceId: string): Map<string, Record<string, string>> {
		return bacnetFieldErrors.get(deviceId) ?? new Map();
	}

	function getBacnetClientErrors(deviceId: string): Map<string, Record<string, string>> {
		return bacnetClientErrors.get(deviceId) ?? new Map();
	}

	return {
		get hasUnsavedChanges() {
			return pendingEdits.size > 0 || pendingBacnetEdits.size > 0;
		},
		get pendingCount() {
			return pendingEdits.size + pendingBacnetEdits.size;
		},
		queueEdit,
		queueSpecEdit,
		isFieldDirty,
		isSpecFieldDirty,
		getPendingValue,
		getPendingSpecValue,
		getFieldError,
		queueBacnetEdit,
		clearBacnetFieldError,
		validateBacnetEdits,
		buildBacnetObjectsPayload,
		saveAllPendingEdits,
		discardAllEdits,
		getBacnetPendingEdits,
		getBacnetFieldErrors,
		getBacnetClientErrors,
		saveDeviceEdits,
		saveDeviceBacnetEdits
	};
}
