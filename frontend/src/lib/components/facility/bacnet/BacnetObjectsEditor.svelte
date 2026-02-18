<script lang="ts">
	/**
	 * BacnetObjectsEditor Component
	 * Editable BACnet objects table for inline editing within expanded field device rows
	 */
	import {
		EditableCell,
		EditableSelectCell,
		EditableBooleanCell
	} from '$lib/components/ui/editable-cell/index.js';
	import {
		BACNET_SOFTWARE_TYPES,
		BACNET_HARDWARE_TYPES
	} from '$lib/domain/facility/bacnet-object.js';
	import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
	import type { BacnetObjectInput } from '$lib/domain/facility/field-device.js';
	import { createTranslator } from '$lib/i18n/translator.js';

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

	const t = createTranslator();

	const softwareTypeOptions = BACNET_SOFTWARE_TYPES.map((t) => ({
		value: t.value,
		label: t.value.toUpperCase()
	}));
	const hardwareTypeOptions = BACNET_HARDWARE_TYPES.map((t) => ({
		value: t.value,
		label: t.value.toUpperCase()
	}));

	function isDirty(objectId: string, field: string): boolean {
		const edits = pendingEdits.get(objectId);
		return edits ? field in edits : false;
	}

	function getPendingTextValue(
		objectId: string,
		field: string,
		originalValue: string
	): string | undefined {
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

	function hasTextIndividual(obj: BacnetObject): boolean {
		const edits = pendingEdits.get(obj.id);
		if (edits && 'text_individual' in edits) {
			return !!edits.text_individual;
		}
		return !!obj.text_individual;
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
					<th class="pr-2 pb-2">{$t('field_device.bacnet.table.text_fix')}</th>
					<th class="pr-2 pb-2">{$t('field_device.bacnet.table.description')}</th>
					<th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.software')}</th>
					<th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.hardware')}</th>
					<th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.gms_visible')}</th>
					<th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.optional')}</th>
					<th class="pr-2 pb-2 text-center">
						{$t('field_device.bacnet.table.text_individual')}
					</th>
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
						<td class="max-w-sm py-1 pr-1">
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
							<div class="flex">
								<EditableSelectCell
									value={obj.software_type}
									options={softwareTypeOptions}
									pendingValue={getPendingTextValue(obj.id, 'software_type', obj.software_type)}
									isDirty={isDirty(obj.id, 'software_type')}
									error={getFieldError(obj.id, 'software_type')}
									{disabled}
									onSave={(v) => onEdit(obj.id, 'software_type', v)}
								/>
								<EditableCell
									value={String(obj.software_number).padStart(2, '0')}
									pendingValue={getPendingTextValue(
										obj.id,
										'software_number',
										String(obj.software_number).padStart(2, '0')
									)}
									type="number"
									min={1}
									max={99}
									isDirty={isDirty(obj.id, 'software_number')}
									error={getFieldError(obj.id, 'software_number')}
									{disabled}
									onSave={(v) => {
										const n = v ? Math.max(1, Math.min(99, parseInt(v))) : 1;
										onEdit(obj.id, 'software_number', n);
									}}
								/>
							</div>
						</td>
						<td class="py-1 pr-1">
							<div class="flex">
								<EditableSelectCell
									value={obj.hardware_type}
									options={hardwareTypeOptions}
									pendingValue={getPendingTextValue(obj.id, 'hardware_type', obj.hardware_type)}
									isDirty={isDirty(obj.id, 'hardware_type')}
									error={getFieldError(obj.id, 'hardware_type')}
									{disabled}
									onSave={(v) => onEdit(obj.id, 'hardware_type', v)}
								/>
								<EditableCell
									value={String(obj.hardware_quantity).padStart(2, '0')}
									pendingValue={getPendingTextValue(
										obj.id,
										'hardware_quantity',
										String(obj.hardware_quantity).padStart(2, '0')
									)}
									type="number"
									min={1}
									max={99}
									isDirty={isDirty(obj.id, 'hardware_quantity')}
									error={getFieldError(obj.id, 'hardware_quantity')}
									{disabled}
									onSave={(v) => {
										const n = v ? Math.max(1, Math.min(99, parseInt(v))) : 1;
										onEdit(obj.id, 'hardware_quantity', n);
									}}
								/>
							</div>
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
						<td class="py-1 pr-1">
							<EditableBooleanCell
								value={obj.optional}
								pendingValue={getPendingBoolValue(obj.id, 'optional')}
								isDirty={isDirty(obj.id, 'optional')}
								error={getFieldError(obj.id, 'optional')}
								{disabled}
								onToggle={(v) => onEdit(obj.id, 'optional', v)}
							/>
						</td>
						<td class="py-1">
							{#if hasTextIndividual(obj)}
								<EditableCell
									value={obj.text_individual || ''}
									pendingValue={getPendingTextValue(
										obj.id,
										'text_individual',
										obj.text_individual || ''
									)}
									maxlength={250}
									isDirty={isDirty(obj.id, 'text_individual')}
									error={getFieldError(obj.id, 'text_individual')}
									{disabled}
									onSave={(v) => onEdit(obj.id, 'text_individual', v || undefined)}
								/>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{:else}
	<p class="text-sm text-muted-foreground italic">
		{$t('field_device.bacnet.empty')}
	</p>
{/if}
