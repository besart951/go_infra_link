<script lang="ts">
	/**
	 * BacnetObjectsEditor Component
	 * Editable BACnet objects table for inline editing within expanded field device rows
	 */
	import { EditableCell, EditableSelectCell, EditableBooleanCell } from '$lib/components/ui/editable-cell/index.js';
	import {
		BACNET_SOFTWARE_TYPES,
		BACNET_HARDWARE_TYPES
	} from '$lib/domain/facility/bacnet-object.js';
	import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
	import type { BacnetObjectInput } from '$lib/domain/facility/field-device.js';

	interface Props {
		bacnetObjects: BacnetObject[];
		pendingEdits: Map<string, Partial<BacnetObjectInput>>;
		fieldErrors: Map<string, Record<string, string>>;
		clientErrors: Map<string, Record<string, string>>;
		disabled?: boolean;
		onEdit: (objectId: string, field: string, value: unknown) => void;
	}

	let {
		bacnetObjects,
		pendingEdits,
		fieldErrors,
		clientErrors,
		disabled = false,
		onEdit
	}: Props = $props();

	const softwareTypeOptions = BACNET_SOFTWARE_TYPES.map((t) => ({
		value: t.value,
		label: t.label
	}));
	const hardwareTypeOptions = BACNET_HARDWARE_TYPES.map((t) => ({
		value: t.value,
		label: t.label
	}));

	function isDirty(objectId: string, field: string): boolean {
		const edits = pendingEdits.get(objectId);
		return edits ? field in edits : false;
	}

	function getPendingTextValue(objectId: string, field: string, originalValue: string): string | undefined {
		const edits = pendingEdits.get(objectId);
		if (!edits || !(field in edits)) return undefined;
		const val = (edits as Record<string, unknown>)[field];
		return val !== undefined ? String(val) : undefined;
	}

	function getPendingBoolValue(objectId: string, field: string): boolean | undefined {
		const edits = pendingEdits.get(objectId);
		if (!edits || !(field in edits)) return undefined;
		return (edits as Record<string, unknown>)[field] as boolean;
	}

	function getFieldError(objectId: string, field: string): string | undefined {
		return fieldErrors.get(objectId)?.[field] || clientErrors.get(objectId)?.[field];
	}
</script>

{#if bacnetObjects.length > 0}
	<div class="overflow-x-auto">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b text-left text-xs text-muted-foreground">
					<th class="pr-2 pb-2">Text Fix</th>
					<th class="pr-2 pb-2">Description</th>
					<th class="pr-2 pb-2">Software Type</th>
					<th class="pr-2 pb-2">Software Nr</th>
					<th class="pr-2 pb-2">Hardware Type</th>
					<th class="pr-2 pb-2">Hardware Qty</th>
					<th class="pr-2 pb-2">GMS Visible</th>
					<th class="pb-2">Optional</th>
				</tr>
			</thead>
			<tbody>
				{#each bacnetObjects as obj, index (obj.id)}
					<tr class="border-b border-purple-100 last:border-0 dark:border-purple-900">
						<td class="py-1 pr-1">
							<EditableCell
								value={obj.text_fix}
								pendingValue={getPendingTextValue(obj.id, 'text_fix', obj.text_fix)}
								maxlength={250}
								isDirty={isDirty(obj.id, 'text_fix')}
								error={getFieldError(obj.id, 'text_fix')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'text_fix', v)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableCell
								value={obj.description || ''}
								pendingValue={getPendingTextValue(obj.id, 'description', obj.description || '')}
								maxlength={250}
								isDirty={isDirty(obj.id, 'description')}
								error={getFieldError(obj.id, 'description')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'description', v || undefined)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableSelectCell
								value={obj.software_type}
								options={softwareTypeOptions}
								pendingValue={getPendingTextValue(obj.id, 'software_type', obj.software_type)}
								isDirty={isDirty(obj.id, 'software_type')}
								error={getFieldError(obj.id, 'software_type')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'software_type', v)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableCell
								value={String(obj.software_number)}
								pendingValue={getPendingTextValue(obj.id, 'software_number', String(obj.software_number))}
								type="number"
								min={0}
								max={65535}
								isDirty={isDirty(obj.id, 'software_number')}
								error={getFieldError(obj.id, 'software_number')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'software_number', v ? parseInt(v) : 0)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableSelectCell
								value={obj.hardware_type}
								options={hardwareTypeOptions}
								pendingValue={getPendingTextValue(obj.id, 'hardware_type', obj.hardware_type)}
								isDirty={isDirty(obj.id, 'hardware_type')}
								error={getFieldError(obj.id, 'hardware_type')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'hardware_type', v)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableCell
								value={String(obj.hardware_quantity)}
								pendingValue={getPendingTextValue(obj.id, 'hardware_quantity', String(obj.hardware_quantity))}
								type="number"
								min={1}
								max={255}
								isDirty={isDirty(obj.id, 'hardware_quantity')}
								error={getFieldError(obj.id, 'hardware_quantity')}
								{disabled}
								onSave={(v) => onEdit(obj.id, 'hardware_quantity', v ? parseInt(v) : 1)}
							/>
						</td>
						<td class="py-1 pr-1">
							<EditableBooleanCell
								value={obj.gms_visible}
								pendingValue={getPendingBoolValue(obj.id, 'gms_visible')}
								isDirty={isDirty(obj.id, 'gms_visible')}
								error={getFieldError(obj.id, 'gms_visible')}
								{disabled}
								onToggle={(v) => onEdit(obj.id, 'gms_visible', v)}
							/>
						</td>
						<td class="py-1">
							<EditableBooleanCell
								value={obj.optional}
								pendingValue={getPendingBoolValue(obj.id, 'optional')}
								isDirty={isDirty(obj.id, 'optional')}
								error={getFieldError(obj.id, 'optional')}
								{disabled}
								onToggle={(v) => onEdit(obj.id, 'optional', v)}
							/>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p class="text-sm text-muted-foreground italic">
		No BACnet objects configured for this field device.
	</p>
{/if}
