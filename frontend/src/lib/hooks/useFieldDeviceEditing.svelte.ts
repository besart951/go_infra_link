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
import { bulkUpdateFieldDevices } from '$lib/infrastructure/api/facility.adapter.js';
import { addToast } from '$lib/components/toast.svelte';
import type {
	FieldDevice,
	UpdateFieldDeviceRequest,
	BulkUpdateFieldDeviceItem,
	SpecificationInput,
	BacnetObjectInput
} from '$lib/domain/facility/index.js';

export interface EditErrorInfo {
	message?: string;
	fields?: Record<string, string>;
}

export function useFieldDeviceEditing() {
	// Pending edits state for inline editing
	let pendingEdits = $state<Map<string, Partial<UpdateFieldDeviceRequest>>>(new Map());

	// BACnet pending edits: deviceId -> (objectId -> partial edits)
	let pendingBacnetEdits = $state<Map<string, Map<string, Partial<BacnetObjectInput>>>>(
		new Map()
	);

	// Error tracking state per device ID
	let editErrors = $state<Map<string, EditErrorInfo>>(new Map());

	// BACnet field errors from server: deviceId -> (objectId -> { field: error })
	let bacnetFieldErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());
	// BACnet client-side validation errors: deviceId -> (objectId -> { field: error })
	let bacnetClientErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());

	function queueEdit(deviceId: string, field: keyof UpdateFieldDeviceRequest, value: unknown) {
		const existing = pendingEdits.get(deviceId) || {};
		pendingEdits = new Map(pendingEdits).set(deviceId, { ...existing, [field]: value });
		// Clear any existing error for this device when editing
		if (editErrors.has(deviceId)) {
			const newErrors = new Map(editErrors);
			newErrors.delete(deviceId);
			editErrors = newErrors;
		}
	}

	function queueSpecEdit(deviceId: string, field: keyof SpecificationInput, value: unknown) {
		const existing = pendingEdits.get(deviceId) || {};
		const existingSpec =
			((existing as Record<string, unknown>)._specification as SpecificationInput) || {};
		const newSpec = { ...existingSpec, [field]: value };
		pendingEdits = new Map(pendingEdits).set(deviceId, {
			...existing,
			_specification: newSpec
		} as Partial<UpdateFieldDeviceRequest>);
		// Clear any existing error
		if (editErrors.has(deviceId)) {
			const newErrors = new Map(editErrors);
			newErrors.delete(deviceId);
			editErrors = newErrors;
		}
	}

	function isFieldDirty(deviceId: string, field: keyof UpdateFieldDeviceRequest): boolean {
		const edit = pendingEdits.get(deviceId);
		return edit ? field in edit : false;
	}

	function isSpecFieldDirty(deviceId: string, field: keyof SpecificationInput): boolean {
		const edit = pendingEdits.get(deviceId);
		if (!edit) return false;
		const spec = (edit as Record<string, unknown>)._specification as
			| SpecificationInput
			| undefined;
		return spec ? field in spec : false;
	}

	function getPendingValue(
		deviceId: string,
		field: keyof UpdateFieldDeviceRequest
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
		const spec = (edit as Record<string, unknown>)._specification as
			| SpecificationInput
			| undefined;
		if (!spec || !(field in spec)) return undefined;
		const val = spec[field];
		return val !== undefined ? String(val) : undefined;
	}

	function getFieldError(deviceId: string, field: string): string | undefined {
		const errorInfo = editErrors.get(deviceId);
		if (!errorInfo) return undefined;

		if (errorInfo.fields && Object.keys(errorInfo.fields).length > 0) {
			if (errorInfo.fields[field]) return errorInfo.fields[field];
			if (errorInfo.fields[`fielddevice.${field}`])
				return errorInfo.fields[`fielddevice.${field}`];
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
	): BacnetObjectInput[] {
		if (!device.bacnet_objects) return [];
		return device.bacnet_objects.map((obj) => {
			const edits = deviceEdits.get(obj.id) || {};
			return {
				text_fix: 'text_fix' in edits ? (edits.text_fix as string) : obj.text_fix,
				description:
					'description' in edits
						? (edits.description as string | undefined)
						: obj.description,
				gms_visible:
					'gms_visible' in edits ? (edits.gms_visible as boolean) : obj.gms_visible,
				optional: 'optional' in edits ? (edits.optional as boolean) : obj.optional,
				software_type:
					'software_type' in edits
						? (edits.software_type as string)
						: obj.software_type,
				software_number:
					'software_number' in edits
						? (edits.software_number as number)
						: obj.software_number,
				hardware_type:
					'hardware_type' in edits
						? (edits.hardware_type as string)
						: obj.hardware_type,
				hardware_quantity:
					'hardware_quantity' in edits
						? (edits.hardware_quantity as number)
						: obj.hardware_quantity,
				software_reference_id: obj.software_reference_id,
				state_text_id: obj.state_text_id,
				notification_class_id: obj.notification_class_id,
				alarm_definition_id: obj.alarm_definition_id
			};
		});
	}

	async function saveAllPendingEdits(
		storeItems: FieldDevice[],
		reloadFn: () => void
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
		for (const id of allDeviceIds) {
			const changes = pendingEdits.get(id) || {};
			const spec = (changes as Record<string, unknown>)._specification as
				| SpecificationInput
				| undefined;

			const update: BulkUpdateFieldDeviceItem = {
				id,
				bmk: changes.bmk,
				description: changes.description,
				apparat_nr: changes.apparat_nr,
				apparat_id: changes.apparat_id,
				system_part_id: changes.system_part_id,
				specification: spec
			};

			// Include BACnet objects if there are pending BACnet edits for this device
			const bacnetEditsForDevice = pendingBacnetEdits.get(id);
			if (bacnetEditsForDevice && bacnetEditsForDevice.size > 0) {
				const device = storeItems.find((d) => d.id === id);
				if (device) {
					update.bacnet_objects = buildBacnetObjectsPayload(device, bacnetEditsForDevice);
				}
			}

			updates.push(update);
		}

		try {
			const result = await bulkUpdateFieldDevices({ updates });

			// Process results and track errors
			const newErrors = new Map<string, EditErrorInfo>();
			const newBacnetFieldErrors = new Map<string, Map<string, Record<string, string>>>();
			const successIds = new Set<string>();

			for (const r of result.results) {
				if (r.success) {
					successIds.add(r.id);
				} else {
					if (r.error) {
						newErrors.set(r.id, { message: r.error, fields: r.fields });
					}
					// Parse field-level errors for BACnet objects
					if (r.fields) {
						const device = storeItems.find((d) => d.id === r.id);
						if (device?.bacnet_objects) {
							const objErrors = new Map<string, Record<string, string>>();
							for (const [fieldPath, msg] of Object.entries(r.fields)) {
								const match = fieldPath.match(/^bacnet_objects\.(\d+)\.(.+)$/);
								if (match) {
									const idx = parseInt(match[1]);
									const field = match[2];
									if (device.bacnet_objects[idx]) {
										const objId = device.bacnet_objects[idx].id;
										const existing = objErrors.get(objId) || {};
										existing[field] = msg;
										objErrors.set(objId, existing);
									}
								}
							}
							if (objErrors.size > 0) {
								newBacnetFieldErrors.set(r.id, objErrors);
							}
						}
					}
				}
			}

			// Remove successful edits from pending, keep failed ones
			const remainingEdits = new Map(pendingEdits);
			const remainingBacnetEdits = new Map(pendingBacnetEdits);
			for (const id of successIds) {
				remainingEdits.delete(id);
				remainingBacnetEdits.delete(id);
			}
			pendingEdits = remainingEdits;
			pendingBacnetEdits = remainingBacnetEdits;
			editErrors = newErrors;
			bacnetFieldErrors = newBacnetFieldErrors;

			if (result.success_count > 0) {
				addToast(`Updated ${result.success_count} field device(s)`, 'success');
				reloadFn();
			}
			if (result.failure_count > 0) {
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

	function discardAllEdits() {
		pendingEdits = new Map();
		pendingBacnetEdits = new Map();
		editErrors = new Map();
		bacnetFieldErrors = new Map();
		bacnetClientErrors = new Map();
	}

	function getBacnetPendingEdits(
		deviceId: string
	): Map<string, Partial<BacnetObjectInput>> {
		return pendingBacnetEdits.get(deviceId) ?? new Map();
	}

	function getBacnetFieldErrors(
		deviceId: string
	): Map<string, Record<string, string>> {
		return bacnetFieldErrors.get(deviceId) ?? new Map();
	}

	function getBacnetClientErrors(
		deviceId: string
	): Map<string, Record<string, string>> {
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
		getBacnetClientErrors
	};
}
