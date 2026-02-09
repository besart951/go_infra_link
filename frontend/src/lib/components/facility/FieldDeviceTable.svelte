<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Settings2 } from '@lucide/svelte';
	import FieldDeviceTableRow from '$lib/components/facility/FieldDeviceTableRow.svelte';
	import BacnetObjectsEditor from '$lib/components/facility/BacnetObjectsEditor.svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';

	interface Props {
		items: FieldDevice[];
		loading: boolean;
		searchInput: string;
		allApparats: Apparat[];
		allSystemParts: SystemPart[];
		selectedIds: Set<string>;
		editing: ReturnType<typeof useFieldDeviceEditing>;
		allSelected: boolean;
		someSelected: boolean;
		onToggleSelect: (id: string) => void;
		onToggleSelectAll: () => void;
		onAutoSave: (updated: FieldDevice) => void;
	}

	let {
		items,
		loading,
		searchInput,
		allApparats,
		allSystemParts,
		selectedIds,
		editing,
		allSelected,
		someSelected,
		onToggleSelect,
		onToggleSelectAll,
		onAutoSave
	}: Props = $props();

	let expandedBacnetRows = $state<Set<string>>(new Set());
	let showSpecifications = $state(false);

	const baseColumnCount = 10;
	const specColumnCount = 11;
	const columnCount = $derived(
		showSpecifications ? baseColumnCount + specColumnCount : baseColumnCount
	);

	function toggleRowExpansion(id: string) {
		const newSet = new Set(expandedBacnetRows);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		expandedBacnetRows = newSet;
	}

	function toggleSpecifications() {
		showSpecifications = !showSpecifications;
	}
</script>

<div class="rounded-lg border bg-background">
	<Table.Root>
		<Table.Header>
			<Table.Row>
				<!-- Selection Checkbox -->
				<Table.Head class="w-10">
					<Checkbox
						checked={allSelected}
						indeterminate={someSelected}
						onCheckedChange={onToggleSelectAll}
						aria-label="Select all"
					/>
				</Table.Head>
				<!-- Expand Column for BACnet Objects -->
				<Table.Head class="w-10"></Table.Head>
				<Table.Head>SPS System Type</Table.Head>
				<Table.Head>BMK</Table.Head>
				<Table.Head>Description</Table.Head>
				<Table.Head class="w-24">Apparat Nr</Table.Head>
				<Table.Head class="w-48">Apparat</Table.Head>
				<Table.Head class="w-48">System Part</Table.Head>
				<!-- Specification Toggle Header -->
				<Table.Head class="w-10">
					<Button
						variant={showSpecifications ? 'secondary' : 'ghost'}
						size="sm"
						class="h-7 w-7 p-0"
						onclick={toggleSpecifications}
						title={showSpecifications ? 'Hide specifications' : 'Show specifications'}
					>
						<Settings2 class="h-4 w-4" />
					</Button>
				</Table.Head>
				{#if showSpecifications}
					<Table.Head class="text-xs">Supplier</Table.Head>
					<Table.Head class="text-xs">Brand</Table.Head>
					<Table.Head class="text-xs">Type</Table.Head>
					<Table.Head class="text-xs">Motor/Valve</Table.Head>
					<Table.Head class="text-xs">Size</Table.Head>
					<Table.Head class="text-xs">Install Loc.</Table.Head>
					<Table.Head class="text-xs">PH</Table.Head>
					<Table.Head class="text-xs">AC/DC</Table.Head>
					<Table.Head class="text-xs">Amperage</Table.Head>
					<Table.Head class="text-xs">Power</Table.Head>
					<Table.Head class="text-xs">Rotation</Table.Head>
				{/if}
				<Table.Head class="w-28">Created</Table.Head>
				<Table.Head class="w-24">Actions</Table.Head>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#if loading && items.length === 0}
				{#each Array(5) as _}
					<Table.Row>
						{#each Array(columnCount) as _}
							<Table.Cell>
								<Skeleton class="h-8 w-full" />
							</Table.Cell>
						{/each}
					</Table.Row>
				{/each}
			{:else if items.length === 0}
				<Table.Row>
					<Table.Cell colspan={columnCount} class="h-24 text-center">
						<div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
							<p class="font-medium">
								No field devices found. Create your first field device to get started.
							</p>
							{#if searchInput}
								<p class="text-sm">Try adjusting your search</p>
							{/if}
						</div>
					</Table.Cell>
				</Table.Row>
			{:else}
				{#each items as device (device.id)}
					<FieldDeviceTableRow
						{device}
						isSelected={selectedIds.has(device.id)}
						{showSpecifications}
						{allApparats}
						{allSystemParts}
						isExpanded={expandedBacnetRows.has(device.id)}
						{loading}
						{editing}
						onToggleSelect={() => onToggleSelect(device.id)}
						onToggleExpansion={() => toggleRowExpansion(device.id)}
						onAutoSave={onAutoSave}
					/>

					<!-- Expanded BACnet Objects Row -->
					{#if expandedBacnetRows.has(device.id)}
						<Table.Row
							class="bg-purple-50/50 hover:bg-purple-50/70 dark:bg-purple-950/20 dark:hover:bg-purple-950/30"
						>
							<Table.Cell colspan={columnCount} class="p-0">
								<div class="border-l-4 border-l-purple-500 py-4 pr-4 pl-14">
									<div class="mb-3 flex items-center gap-2">
										<span
											class="rounded-full bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700 dark:bg-purple-900 dark:text-purple-300"
										>
											BACnet Objects
										</span>
										<span class="text-xs text-muted-foreground">
											{device.bacnet_objects?.length || 0} object(s)
										</span>
									</div>
										<BacnetObjectsEditor
										bacnetObjects={device.bacnet_objects ?? []}
										pendingEdits={editing.getBacnetPendingEdits(device.id)}
										fieldErrors={editing.getBacnetFieldErrors(device.id)}
										clientErrors={editing.getBacnetClientErrors(device.id)}
											onEdit={(objectId, field, value) => {
												editing.queueBacnetEdit(device.id, objectId, field, value);
												editing.saveDeviceBacnetEdits(device, onAutoSave);
											}}
									/>
								</div>
							</Table.Cell>
						</Table.Row>
					{/if}
				{/each}
			{/if}
		</Table.Body>
	</Table.Root>
</div>
