<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { ChevronDown, ChevronRight, CopyCheck, Eraser } from '@lucide/svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import TableApparatSelect from '$lib/components/facility/TableApparatSelect.svelte';
	import TableSystemPartSelect from '$lib/components/facility/TableSystemPartSelect.svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import type {
		Apparat,
		SystemPart,
		UpdateFieldDeviceRequest,
		SpecificationInput
	} from '$lib/domain/facility/index.js';

	interface Props {
		selectedCount: number;
		selectedIds: Set<string>;
		allApparats: Apparat[];
		allSystemParts: SystemPart[];
		editing: ReturnType<typeof useFieldDeviceEditing>;
	}

	let { selectedCount, selectedIds, allApparats, allSystemParts, editing }: Props = $props();

	let bulkEditValues = $state<Partial<UpdateFieldDeviceRequest>>({});
	let bulkSpecValues = $state<Partial<SpecificationInput>>({});
	let showBulkSpecFields = $state(false);

	const hasBulkValues = $derived(
		Object.values(bulkEditValues).some((v) => v !== undefined && v !== '') ||
			Object.values(bulkSpecValues).some((v) => v !== undefined && v !== '')
	);

	function applyBulkEdits() {
		if (selectedIds.size === 0) return;

		let appliedCount = 0;
		for (const deviceId of selectedIds) {
			for (const [field, value] of Object.entries(bulkEditValues)) {
				if (value !== undefined && value !== '') {
					editing.queueEdit(deviceId, field as keyof UpdateFieldDeviceRequest, value);
					appliedCount++;
				}
			}
			for (const [field, value] of Object.entries(bulkSpecValues)) {
				if (value !== undefined && value !== '') {
					editing.queueSpecEdit(deviceId, field as keyof SpecificationInput, value);
					appliedCount++;
				}
			}
		}

		if (appliedCount > 0) {
			addToast(`Applied edits to ${selectedIds.size} device(s)`, 'success');
		} else {
			addToast('No fields filled in to apply', 'error');
		}
	}

	function clearBulkEdit() {
		bulkEditValues = {};
		bulkSpecValues = {};
	}
</script>

<Card.Root class="border-primary/30 bg-primary/5">
	<Card.Header class="pb-3">
		<Card.Title class="text-base">Bulk Edit ({selectedCount} selected)</Card.Title>
		<Card.Description>
			Fill in fields and click "Apply to Selected" to queue changes. Then "Save All" to persist.
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<!-- Device Fields -->
		<div class="mb-4">
			<h4 class="mb-2 text-sm font-medium">Device Fields</h4>
			<div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-5">
				<div class="flex flex-col gap-1">
					<Label class="text-xs">BMK</Label>
					<Input
						type="text"
						placeholder="BMK"
						maxlength={10}
						value={bulkEditValues.bmk ?? ''}
						oninput={(e: Event) => {
							const v = (e.target as HTMLInputElement).value;
							bulkEditValues = { ...bulkEditValues, bmk: v || undefined };
						}}
					/>
				</div>
				<div class="flex flex-col gap-1">
					<Label class="text-xs">Description</Label>
					<Input
						type="text"
						placeholder="Description"
						maxlength={250}
						value={bulkEditValues.description ?? ''}
						oninput={(e: Event) => {
							const v = (e.target as HTMLInputElement).value;
							bulkEditValues = { ...bulkEditValues, description: v || undefined };
						}}
					/>
				</div>
				<div class="flex flex-col gap-1">
					<Label class="text-xs">Apparat Nr</Label>
					<Input
						type="number"
						placeholder="1-99"
						min={1}
						max={99}
						value={bulkEditValues.apparat_nr?.toString() ?? ''}
						oninput={(e: Event) => {
							const v = (e.target as HTMLInputElement).value;
							bulkEditValues = {
								...bulkEditValues,
								apparat_nr: v ? parseInt(v) : undefined
							};
						}}
					/>
				</div>
				<div class="flex flex-col gap-1">
					<Label class="text-xs">Apparat</Label>
					<TableApparatSelect
						items={allApparats}
						value={bulkEditValues.apparat_id ?? ''}
						width="w-full"
						onValueChange={(v) => {
							bulkEditValues = { ...bulkEditValues, apparat_id: v || undefined };
						}}
					/>
				</div>
				<div class="flex flex-col gap-1">
					<Label class="text-xs">System Part</Label>
					<TableSystemPartSelect
						items={allSystemParts}
						value={bulkEditValues.system_part_id ?? ''}
						width="w-full"
						onValueChange={(v) => {
							bulkEditValues = { ...bulkEditValues, system_part_id: v || undefined };
						}}
					/>
				</div>
			</div>
		</div>

		<!-- Specification Fields (collapsible) -->
		<div class="mb-4">
			<button
				type="button"
				class="mb-2 flex items-center gap-1 text-sm font-medium hover:underline"
				onclick={() => (showBulkSpecFields = !showBulkSpecFields)}
			>
				{#if showBulkSpecFields}
					<ChevronDown class="h-4 w-4" />
				{:else}
					<ChevronRight class="h-4 w-4" />
				{/if}
				Specification Fields
			</button>
			{#if showBulkSpecFields}
				<div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Supplier</Label>
						<Input
							type="text"
							placeholder="Supplier"
							maxlength={250}
							value={bulkSpecValues.specification_supplier ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									specification_supplier: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Brand</Label>
						<Input
							type="text"
							placeholder="Brand"
							maxlength={250}
							value={bulkSpecValues.specification_brand ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									specification_brand: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Type</Label>
						<Input
							type="text"
							placeholder="Type"
							maxlength={250}
							value={bulkSpecValues.specification_type ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									specification_type: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Motor/Valve</Label>
						<Input
							type="text"
							placeholder="Motor/Valve"
							maxlength={250}
							value={bulkSpecValues.additional_info_motor_valve ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									additional_info_motor_valve: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Size</Label>
						<Input
							type="number"
							placeholder="Size"
							value={bulkSpecValues.additional_info_size?.toString() ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									additional_info_size: v ? parseInt(v) : undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Install Location</Label>
						<Input
							type="text"
							placeholder="Install Location"
							maxlength={250}
							value={bulkSpecValues.additional_information_installation_location ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									additional_information_installation_location: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">PH</Label>
						<Input
							type="number"
							placeholder="PH"
							value={bulkSpecValues.electrical_connection_ph?.toString() ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									electrical_connection_ph: v ? parseInt(v) : undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">AC/DC</Label>
						<Input
							type="text"
							placeholder="AC/DC"
							maxlength={2}
							value={bulkSpecValues.electrical_connection_acdc ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									electrical_connection_acdc: v || undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Amperage</Label>
						<Input
							type="number"
							placeholder="A"
							value={bulkSpecValues.electrical_connection_amperage?.toString() ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									electrical_connection_amperage: v ? parseFloat(v) : undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Power</Label>
						<Input
							type="number"
							placeholder="W"
							value={bulkSpecValues.electrical_connection_power?.toString() ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									electrical_connection_power: v ? parseFloat(v) : undefined
								};
							}}
						/>
					</div>
					<div class="flex flex-col gap-1">
						<Label class="text-xs">Rotation</Label>
						<Input
							type="number"
							placeholder="RPM"
							value={bulkSpecValues.electrical_connection_rotation?.toString() ?? ''}
							oninput={(e: Event) => {
								const v = (e.target as HTMLInputElement).value;
								bulkSpecValues = {
									...bulkSpecValues,
									electrical_connection_rotation: v ? parseInt(v) : undefined
								};
							}}
						/>
					</div>
				</div>
			{/if}
		</div>

		<!-- Actions -->
		<div class="flex gap-2">
			<Button size="sm" onclick={applyBulkEdits} disabled={!hasBulkValues}>
				<CopyCheck class="mr-1 h-4 w-4" />
				Apply to Selected
			</Button>
			<Button variant="outline" size="sm" onclick={clearBulkEdit} disabled={!hasBulkValues}>
				<Eraser class="mr-1 h-4 w-4" />
				Clear
			</Button>
		</div>
	</Card.Content>
</Card.Root>
