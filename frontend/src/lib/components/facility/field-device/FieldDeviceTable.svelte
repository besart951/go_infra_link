<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { ArrowDown, ArrowUp, Settings2 } from '@lucide/svelte';
	import FieldDeviceTableRow from './FieldDeviceTableRow.svelte';
	import BacnetObjectsEditor from '../bacnet/BacnetObjectsEditor.svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

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
		onCopy: (value: string) => void;
		onDelete: (device: FieldDevice) => void;
		sortBy?: string;
		sortOrder?: 'asc' | 'desc';
		onSort: (orderBy: string) => void;
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
		onCopy,
		onDelete,
		sortBy,
		sortOrder,
		onSort
	}: Props = $props();

	let expandedBacnetRows = $state<Set<string>>(new Set());
	let showSpecifications = $state(false);

	const t = createTranslator();

	const baseColumnCount = 11;
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

	function sortState(key: string) {
		if (!sortBy || sortBy !== key) return undefined;
		return sortOrder === 'desc' ? 'desc' : 'asc';
	}
</script>

<div class="rounded-lg border bg-background">
	<Table.Root class="[&_td]:p-2 [&_th]:h-10 [&_th]:px-2">
		<Table.Header>
			<Table.Row>
				<!-- Selection Checkbox -->
				<Table.Head class="w-10">
					<Checkbox
						checked={allSelected}
						indeterminate={someSelected}
						onCheckedChange={onToggleSelectAll}
						aria-label={$t('field_device.table.select_all')}
					/>
				</Table.Head>
				<!-- Expand Column for BACnet Objects -->
				<Table.Head class="w-10"></Table.Head>
				<Table.Head>
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 text-left underline-offset-4 hover:underline"
						onclick={() => onSort('sps_system_type')}
					>
						<span>{$t('field_device.table.sps_system_type')}</span>
						{#if sortState('sps_system_type') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('sps_system_type') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head>
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('bmk')}
					>
						<span>{$t('field_device.table.bmk')}</span>
						{#if sortState('bmk') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('bmk') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head>
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('description')}
					>
						<span>{$t('field_device.table.description')}</span>
						{#if sortState('description') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('description') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head>
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('text_fix')}
					>
						<span>{$t('field_device.table.text_fix')}</span>
						{#if sortState('text_fix') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('text_fix') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head class="w-24">
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('apparat_nr')}
					>
						<span>{$t('field_device.table.apparat_nr')}</span>
						{#if sortState('apparat_nr') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('apparat_nr') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head class="w-48">
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('apparat')}
					>
						<span>{$t('field_device.table.apparat')}</span>
						{#if sortState('apparat') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('apparat') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<Table.Head class="w-48">
					<button
						type="button"
						class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
						onclick={() => onSort('system_part')}
					>
						<span>{$t('field_device.table.system_part')}</span>
						{#if sortState('system_part') === 'asc'}
							<ArrowUp class="h-3 w-3" />
						{:else if sortState('system_part') === 'desc'}
							<ArrowDown class="h-3 w-3" />
						{/if}
					</button>
				</Table.Head>
				<!-- Specification Toggle Header -->
				<Table.Head class="w-10">
					<Button
						variant={showSpecifications ? 'secondary' : 'ghost'}
						size="sm"
						class="h-7 w-7 p-0"
						onclick={toggleSpecifications}
						title={showSpecifications
							? $t('field_device.table.hide_specifications')
							: $t('field_device.table.show_specifications')}
					>
						<Settings2 class="h-4 w-4" />
					</Button>
				</Table.Head>
				{#if showSpecifications}
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_supplier')}
						>
							<span>{$t('field_device.table.supplier')}</span>
							{#if sortState('spec_supplier') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_supplier') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_brand')}
						>
							<span>{$t('field_device.table.brand')}</span>
							{#if sortState('spec_brand') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_brand') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_type')}
						>
							<span>{$t('field_device.table.type')}</span>
							{#if sortState('spec_type') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_type') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_motor_valve')}
						>
							<span>{$t('field_device.table.motor_valve')}</span>
							{#if sortState('spec_motor_valve') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_motor_valve') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_size')}
						>
							<span>{$t('field_device.table.size')}</span>
							{#if sortState('spec_size') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_size') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_install_loc')}
						>
							<span>{$t('field_device.table.install_location')}</span>
							{#if sortState('spec_install_loc') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_install_loc') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_ph')}
						>
							<span>{$t('field_device.table.ph')}</span>
							{#if sortState('spec_ph') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_ph') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_acdc')}
						>
							<span>{$t('field_device.table.acdc')}</span>
							{#if sortState('spec_acdc') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_acdc') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_amperage')}
						>
							<span>{$t('field_device.table.amperage')}</span>
							{#if sortState('spec_amperage') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_amperage') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_power')}
						>
							<span>{$t('field_device.table.power')}</span>
							{#if sortState('spec_power') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_power') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-xs">
						<button
							type="button"
							class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
							onclick={() => onSort('spec_rotation')}
						>
							<span>{$t('field_device.table.rotation')}</span>
							{#if sortState('spec_rotation') === 'asc'}
								<ArrowUp class="h-3 w-3" />
							{:else if sortState('spec_rotation') === 'desc'}
								<ArrowDown class="h-3 w-3" />
							{/if}
						</button>
					</Table.Head>
				{/if}
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
							<p class="font-medium">{$t('field_device.empty.title')}</p>
							{#if searchInput}
								<p class="text-sm">{$t('field_device.empty.search_hint')}</p>
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
						{onCopy}
						{onDelete}
						onToggleSelect={() => onToggleSelect(device.id)}
						onToggleExpansion={() => toggleRowExpansion(device.id)}
					/>

					<!-- Expanded BACnet Objects Row -->
					{#if expandedBacnetRows.has(device.id)}
						<Table.Row
							class="bg-purple-50/50 hover:bg-purple-50/70 dark:bg-purple-950/20 dark:hover:bg-purple-950/30"
						>
							<Table.Cell colspan={columnCount} class="p-0">
								<BacnetObjectsEditor
									bacnetObjects={device.bacnet_objects ?? []}
									pendingEdits={editing.getBacnetPendingEdits(device.id) ?? new Map()}
									fieldErrors={editing.getBacnetFieldErrors(device.id) ?? new Map()}
									clientErrors={editing.getBacnetClientErrors(device.id) ?? new Map()}
									onEdit={(objectId, field, value) => {
										editing.queueBacnetEdit(device.id, objectId, field, value);
									}}
								/>
							</Table.Cell>
						</Table.Row>
					{/if}
				{/each}
			{/if}
		</Table.Body>
	</Table.Root>
</div>
