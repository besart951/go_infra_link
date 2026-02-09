<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
	import { ChevronDown, ChevronRight } from '@lucide/svelte';
	import TableApparatSelect from '$lib/components/facility/TableApparatSelect.svelte';
	import TableSystemPartSelect from '$lib/components/facility/TableSystemPartSelect.svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';

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
		onAutoSave: (updated: FieldDevice) => void;
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
		onAutoSave
	}: Props = $props();

	function formatSPSControllerSystemType(dev: FieldDevice): string {
		const sysType = dev.sps_controller_system_type;
		if (!sysType) return '-';
		const number = sysType.number ?? '';
		const docName = sysType.document_name ?? '';
		if (number && docName) return `${number} - ${docName}`;
		if (number) return String(number);
		if (docName) return docName;
		return '-';
	}

	function handleApparatChange(newApparatId: string) {
		if (!newApparatId || newApparatId === device.apparat_id) return;
		editing.queueEdit(device.id, 'apparat_id', newApparatId);
		editing.saveDeviceEdits(device, onAutoSave);
	}

	function handleSystemPartChange(newSystemPartId: string) {
		if (!newSystemPartId || newSystemPartId === device.system_part_id) return;
		editing.queueEdit(device.id, 'system_part_id', newSystemPartId);
		editing.saveDeviceEdits(device, onAutoSave);
	}
</script>

<Table.Row
	class={[loading ? 'opacity-60' : '', isSelected ? 'bg-muted/50' : '']
		.filter(Boolean)
		.join(' ')}
>
	<!-- Selection Checkbox -->
	<Table.Cell class="p-2">
		<Checkbox
			checked={isSelected}
			onCheckedChange={onToggleSelect}
			aria-label={`Select ${device.bmk || device.id}`}
		/>
	</Table.Cell>
	<!-- Expand button for BACnet Objects -->
	<Table.Cell class="p-2">
		<Button
			variant="ghost"
			size="sm"
			class="h-6 w-6 p-0"
			onclick={onToggleExpansion}
			title="Expand BACnet objects"
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
	<!-- BMK -->
	<Table.Cell class="p-1">
		<EditableCell
			value={device.bmk ?? ''}
			pendingValue={editing.getPendingValue(device.id, 'bmk')}
			type="text"
			maxlength={10}
			isDirty={editing.isFieldDirty(device.id, 'bmk')}
			error={editing.getFieldError(device.id, 'bmk')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'bmk', v || undefined);
				editing.saveDeviceEdits(device, onAutoSave);
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
			isDirty={editing.isFieldDirty(device.id, 'description')}
			error={editing.getFieldError(device.id, 'description')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'description', v || undefined);
				editing.saveDeviceEdits(device, onAutoSave);
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
			isDirty={editing.isFieldDirty(device.id, 'apparat_nr')}
			error={editing.getFieldError(device.id, 'apparat_nr')}
			onSave={(v) => {
				editing.queueEdit(device.id, 'apparat_nr', v ? parseInt(v) : undefined);
				editing.saveDeviceEdits(device, onAutoSave);
			}}
		/>
	</Table.Cell>
	<!-- Apparat (static select with preloaded data) -->
	<Table.Cell>
		<TableApparatSelect
			items={allApparats}
			value={device.apparat_id}
			width="w-full"
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
			error={editing.getFieldError(device.id, 'system_part_id')}
			onValueChange={(newVal) => handleSystemPartChange(newVal)}
		/>
	</Table.Cell>
	<!-- Specification indicator -->
	<Table.Cell class="text-center">
		{#if device.specification}
			<span
				class="inline-block h-2 w-2 rounded-full bg-green-500"
				title="Has specification"
			></span>
		{:else}
			<span
				class="inline-block h-2 w-2 rounded-full bg-gray-300"
				title="No specification"
			></span>
		{/if}
	</Table.Cell>
	<!-- Specification columns (shown when toggled) -->
	{#if showSpecifications}
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.specification_supplier || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_supplier')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_supplier')}
				error={editing.getFieldError(device.id, 'specification_supplier')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_supplier', v || undefined);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.specification_brand || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_brand')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_brand')}
				error={editing.getFieldError(device.id, 'specification_brand')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_brand', v || undefined);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.specification_type || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'specification_type')}
				isDirty={editing.isSpecFieldDirty(device.id, 'specification_type')}
				error={editing.getFieldError(device.id, 'specification_type')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'specification_type', v || undefined);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.additional_info_motor_valve || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'additional_info_motor_valve')}
				isDirty={editing.isSpecFieldDirty(device.id, 'additional_info_motor_valve')}
				error={editing.getFieldError(device.id, 'additional_info_motor_valve')}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'additional_info_motor_valve', v || undefined);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.additional_info_size?.toString() || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'additional_info_size')}
				isDirty={editing.isSpecFieldDirty(device.id, 'additional_info_size')}
				error={editing.getFieldError(device.id, 'additional_info_size')}
				type="number"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'additional_info_size',
						v ? parseInt(v) : undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.additional_information_installation_location || ''}
				pendingValue={editing.getPendingSpecValue(
					device.id,
					'additional_information_installation_location'
				)}
				isDirty={editing.isSpecFieldDirty(
					device.id,
					'additional_information_installation_location'
				)}
				error={editing.getFieldError(
					device.id,
					'additional_information_installation_location'
				)}
				maxlength={250}
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'additional_information_installation_location',
						v || undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.electrical_connection_ph?.toString() || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_ph')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_ph')}
				error={editing.getFieldError(device.id, 'electrical_connection_ph')}
				type="number"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_ph',
						v ? parseInt(v) : undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.electrical_connection_acdc || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_acdc')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_acdc')}
				error={editing.getFieldError(device.id, 'electrical_connection_acdc')}
				maxlength={2}
				placeholder="AC/DC"
				onSave={(v) => {
					editing.queueSpecEdit(device.id, 'electrical_connection_acdc', v || undefined);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.electrical_connection_amperage?.toString() || ''}
				pendingValue={editing.getPendingSpecValue(
					device.id,
					'electrical_connection_amperage'
				)}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_amperage')}
				error={editing.getFieldError(device.id, 'electrical_connection_amperage')}
				type="number"
				placeholder="A"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_amperage',
						v ? parseFloat(v) : undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.electrical_connection_power?.toString() || ''}
				pendingValue={editing.getPendingSpecValue(device.id, 'electrical_connection_power')}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_power')}
				error={editing.getFieldError(device.id, 'electrical_connection_power')}
				type="number"
				placeholder="W"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_power',
						v ? parseFloat(v) : undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
		<Table.Cell class="text-xs">
			<EditableCell
				value={device.specification?.electrical_connection_rotation?.toString() || ''}
				pendingValue={editing.getPendingSpecValue(
					device.id,
					'electrical_connection_rotation'
				)}
				isDirty={editing.isSpecFieldDirty(device.id, 'electrical_connection_rotation')}
				error={editing.getFieldError(device.id, 'electrical_connection_rotation')}
				type="number"
				placeholder="RPM"
				onSave={(v) => {
					editing.queueSpecEdit(
						device.id,
						'electrical_connection_rotation',
						v ? parseInt(v) : undefined
					);
					editing.saveDeviceEdits(device, onAutoSave);
				}}
			/>
		</Table.Cell>
	{/if}
	<!-- Created -->
	<Table.Cell>
		{new Date(device.created_at).toLocaleDateString()}
	</Table.Cell>
	<!-- Actions -->
	<Table.Cell>
		<Button variant="ghost" size="sm">View</Button>
	</Table.Cell>
</Table.Row>
