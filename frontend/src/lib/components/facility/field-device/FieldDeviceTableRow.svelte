<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { ChevronDown, ChevronRight } from '@lucide/svelte';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import TableApparatSelect from '../table-selects/TableApparatSelect.svelte';
	import TableSystemPartSelect from '../table-selects/TableSystemPartSelect.svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';
	import { canPerform } from '$lib/utils/permissions.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		device: FieldDevice;
		isSelected: boolean;
		showSpecifications: boolean;
		allApparats: Apparat[];
		allSystemParts: SystemPart[];
		isExpanded: boolean;
		loading: boolean;
		editing: ReturnType<typeof useFieldDeviceEditing>;
		onToggleSelect: () => void;
		onToggleExpansion: () => void;
		onCopy: (value: string) => void;
		onDelete: (device: FieldDevice) => void;
	}

	let {
		device,
		isSelected,
		showSpecifications,
		allApparats,
		allSystemParts,
		isExpanded,
		loading,
		editing,
		onToggleSelect,
		onToggleExpansion,
		onCopy,
		onDelete
	}: Props = $props();

	const t = createTranslator();

	// Convert specification field values to display strings, handling null/undefined gracefully
	function toDisplayString(value: any, isNumeric = false): string {
		if (value === null || value === undefined || value === '') return '';
		if (isNumeric && typeof value === 'number') return String(value);
		return String(value);
	}

	function formatSPSControllerSystemType(dev: FieldDevice): string {
		const sysType = dev.sps_controller_system_type;
		if (!sysType) return '-';
		
		const deviceName = sysType.sps_controller_name ?? '';
		const number =
			sysType.number === null || sysType.number === undefined
				? ''
				: String(sysType.number).padStart(4, '0');
		const docName = sysType.document_name ?? '';
		
		// Build the system type part
		let sysTypePart = '';
		if (number && docName) {
			sysTypePart = `${number} - ${docName}`;
		} else if (number) {
			sysTypePart = String(number);
		} else if (docName) {
			sysTypePart = docName;
		}
		
		// Combine device name with system type part
		if (deviceName && sysTypePart) return `${deviceName}_${sysTypePart}`;
		if (deviceName) return deviceName;
		if (sysTypePart) return sysTypePart;
		return '-';
	}

	function handleApparatChange(newApparatId: string) {
		if (!newApparatId || newApparatId === device.apparat_id) return;
		editing.queueEdit(device.id, 'apparat_id', newApparatId);
	}

	function handleSystemPartChange(newSystemPartId: string) {
		if (!newSystemPartId || newSystemPartId === device.system_part_id) return;
		editing.queueEdit(device.id, 'system_part_id', newSystemPartId);
	}
</script>

<Table.Row
	class={[loading ? 'opacity-60' : '', isSelected ? 'bg-muted/50' : ''].filter(Boolean).join(' ')}
>
	<!-- Selection Checkbox -->
	<Table.Cell class="p-2">
		<Checkbox
			checked={isSelected}
			onCheckedChange={onToggleSelect}
			aria-label={$t('field_device.table.select_aria', { label: device.bmk || device.id })}
		/>
	</Table.Cell>
	<!-- Expand button for BACnet Objects -->
	<Table.Cell class="p-2">
		<Button
			variant="ghost"
			size="sm"
			class="h-6 w-6 p-0"
			onclick={onToggleExpansion}
			title={$t('field_device.table.bacnet_expand')}
		>
			{#if isExpanded}
				<ChevronDown class="h-4 w-4" />
			{:else}
				<ChevronRight class="h-4 w-4" />
			{/if}
		</Button>
	</Table.Cell>
	<!-- SPS Controller System Type -->
	<Table.Cell class="font-medium">
		{formatSPSControllerSystemType(device)}
	</Table.Cell>
	<Table.Cell class="p-1">
		<EditableCell
			value={device.bmk ?? ''}
			pendingValue={editing.getPendingValue(device.id, 'bmk')}
			type="text"
			maxlength={10}
			disabled={!canPerform('update', 'fielddevice')}
			isDirty={editing.isFieldDirty(device.id, 'bmk')}
			error={editing.getFieldError(device.id, 'bmk')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'bmk', v || undefined);
			}}
		/>
	</Table.Cell>
	<!-- Description -->
	<Table.Cell class="max-w-48 p-1">
		<EditableCell
			value={device.description ?? ''}
			pendingValue={editing.getPendingValue(device.id, 'description')}
			type="text"
			maxlength={250}
			disabled={!canPerform('update', 'fielddevice')}
			isDirty={editing.isFieldDirty(device.id, 'description')}
			error={editing.getFieldError(device.id, 'description')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'description', v || undefined);
			}}
		/>
	</Table.Cell>
	<!-- TextFix -->
	<Table.Cell class="p-1">
		<EditableCell
			value={device.text_fix ?? ''}
			pendingValue={editing.getPendingValue(device.id, 'text_fix')}
			type="text"
			maxlength={250}
			disabled={!canPerform('update', 'fielddevice')}
			isDirty={editing.isFieldDirty(device.id, 'text_fix')}
			error={editing.getFieldError(device.id, 'text_fix')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'text_fix', v || undefined);
			}}
		/>
	</Table.Cell>
	<!-- Apparat Nr -->
	<Table.Cell class="p-1">
		<EditableCell
			value={device.apparat_nr}
			pendingValue={editing.getPendingValue(device.id, 'apparat_nr')}
			type="number"
			min={1}
			max={99}
			disabled={!canPerform('update', 'fielddevice')}
			isDirty={editing.isFieldDirty(device.id, 'apparat_nr')}
			error={editing.getFieldError(device.id, 'apparat_nr')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'apparat_nr', v ? parseInt(v) : undefined);
			}}
		/>
	</Table.Cell>
	<!-- Apparat (static select with preloaded data) -->
	<Table.Cell>
		<TableApparatSelect
			items={allApparats}
			value={device.apparat_id}
			width="w-full"
			disabled={!canPerform('update', 'fielddevice')}
			error={editing.getFieldError(device.id, 'apparat_id')}
			onValueChange={(newVal) => handleApparatChange(newVal)}
		/>
	</Table.Cell>
	<!-- System Part (static select with preloaded data) -->
	<Table.Cell>
		<TableSystemPartSelect
			items={allSystemParts}
			value={device.system_part_id || ''}
			width="w-full"
			disabled={!canPerform('update', 'fielddevice')}
			error={editing.getFieldError(device.id, 'system_part_id')}
			onValueChange={(newVal) => handleSystemPartChange(newVal)}
		/>
	</Table.Cell>
	<!-- Specification indicator -->
	<Table.Cell class="text-center">
		{#if device.specification}
			<span
				class="inline-block h-2 w-2 rounded-full bg-green-500"
				title={$t('field_device.table.spec_available')}
			></span>
		{:else}
			<span
				class="inline-block h-2 w-2 rounded-full bg-gray-300"
				title={$t('field_device.table.spec_missing')}
			></span>
		{/if}
	</Table.Cell>
	<!-- Specification columns (shown when toggled) -->
	{#if showSpecifications}
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.specification_supplier)}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_supplier')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_supplier')}
				error={editing.getFieldError(device.id, 'specification_supplier')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_supplier', v === '' ? null : v);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.specification_brand)}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_brand')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_brand')}
				error={editing.getFieldError(device.id, 'specification_brand')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_brand', v === '' ? null : v);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.specification_type)}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_type')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_type')}
				error={editing.getFieldError(device.id, 'specification_type')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_type', v === '' ? null : v);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.additional_info_motor_valve)}
				pendingValue={editing.getPendingSpecValue(device.id, 'additional_info_motor_valve')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'additional_info_motor_valve')}
				error={editing.getFieldError(device.id, 'additional_info_motor_valve')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'additional_info_motor_valve', v === '' ? null : v);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.additional_info_size, true)}
				pendingValue={editing.getPendingSpecValue(device.id, 'additional_info_size')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'additional_info_size')}
				error={editing.getFieldError(device.id, 'additional_info_size')}
				type="number"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'additional_info_size',
						v === '' ? null : v ? parseInt(v) : null
					);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.additional_information_installation_location)}
				pendingValue={editing.getPendingSpecValue(
					device.id,
					'additional_information_installation_location'
				)}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(
					device.id,
					'additional_information_installation_location'
				)}
				error={editing.getFieldError(device.id, 'additional_information_installation_location')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'additional_information_installation_location',
						v === '' ? null : v
					);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.electrical_connection_ph, true)}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_ph')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_ph')}
				error={editing.getFieldError(device.id, 'electrical_connection_ph')}
				type="number"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_ph',
						v === '' ? null : v ? parseInt(v) : null
					);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.electrical_connection_acdc)}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_acdc')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_acdc')}
				error={editing.getFieldError(device.id, 'electrical_connection_acdc')}
				maxlength={2}
				placeholder={$t('field_device.table.acdc_placeholder')}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'electrical_connection_acdc', v === '' ? null : v);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.electrical_connection_amperage, true)}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_amperage')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_amperage')}
				error={editing.getFieldError(device.id, 'electrical_connection_amperage')}
				type="number"
				placeholder={$t('field_device.table.amperage_placeholder')}
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_amperage',
						v === '' ? null : v ? parseFloat(v) : null
					);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.electrical_connection_power, true)}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_power')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_power')}
				error={editing.getFieldError(device.id, 'electrical_connection_power')}
				type="number"
				placeholder={$t('field_device.table.power_placeholder')}
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_power',
						v === '' ? null : v ? parseFloat(v) : null
					);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={toDisplayString(device.specification?.electrical_connection_rotation, true)}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_rotation')}
				disabled={!canPerform('update', 'fielddevice')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_rotation')}
				error={editing.getFieldError(device.id, 'electrical_connection_rotation')}
				type="number"
				placeholder={$t('field_device.table.rotation_placeholder')}
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_rotation',
						v === '' ? null : v ? parseInt(v) : null
					);
				}}
			/>
		</Table.Cell>
	{/if}
	<!-- Actions -->
	<Table.Cell class="text-right">
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Button variant="ghost" size="icon" {...props}>
						<EllipsisIcon class="size-4" />
					</Button>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content align="end" class="w-40">
				<DropdownMenu.Item
					onclick={() =>
						onCopy(
							device.bmk?.trim() || (device.apparat_nr ? String(device.apparat_nr) : device.id)
						)}
				>
					{$t('facility.copy')}
				</DropdownMenu.Item>
				{#if canPerform('delete', 'fielddevice')}
				<DropdownMenu.Separator />
				<DropdownMenu.Item variant="destructive" onclick={() => onDelete(device)}>
					{$t('common.delete')}
				</DropdownMenu.Item>
				{/if}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Table.Cell>
</Table.Row>
