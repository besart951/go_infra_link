<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
	import {
		Plus,
		X,
		ListPlus,
		ChevronDown,
		ChevronRight,
		Search,
		ChevronLeft,
		Trash2,
		Settings2,
		Save,
		Undo
	} from '@lucide/svelte';
	import { fieldDeviceStore } from '$lib/stores/facility/fieldDeviceStore.js';
	import { lookupCache } from '$lib/stores/facility/lookupCache.js';
	import {
		updateFieldDevice,
		bulkDeleteFieldDevices,
		bulkUpdateFieldDevices
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { addToast } from '$lib/components/toast.svelte';
	import type {
		FieldDevice,
		UpdateFieldDeviceRequest,
		BulkUpdateFieldDeviceItem,
		SpecificationInput
	} from '$lib/domain/facility/index.js';
	import BuildingSelect from '$lib/components/facility/BuildingSelect.svelte';
	import ControlCabinetSelect from '$lib/components/facility/ControlCabinetSelect.svelte';
	import SPSControllerSelect from '$lib/components/facility/SPSControllerSelect.svelte';
	import SPSControllerSystemTypeSelect from '$lib/components/facility/SPSControllerSystemTypeSelect.svelte';
	import FieldDeviceMultiCreateForm from '$lib/components/facility/FieldDeviceMultiCreateForm.svelte';
	import CachedApparatSelect from '$lib/components/facility/CachedApparatSelect.svelte';
	import CachedSystemPartSelect from '$lib/components/facility/CachedSystemPartSelect.svelte';

	// Filter state
	let buildingId = $state('');
	let controlCabinetId = $state('');
	let spsControllerId = $state('');
	let spsControllerSystemTypeId = $state('');
	let showMultiCreateForm = $state(false);
	let expandedBacnetRows = $state<Set<string>>(new Set());
	let showSpecifications = $state(false);
	let searchInput = $state('');

	// Selection state
	let selectedIds = $state<Set<string>>(new Set());

	// Pending edits state for inline editing
	let pendingEdits = $state<Map<string, Partial<UpdateFieldDeviceRequest>>>(new Map());

	// Error tracking state per device ID
	let editErrors = $state<Map<string, string>>(new Map());

	// Derived: has unsaved changes
	const hasUnsavedChanges = $derived(pendingEdits.size > 0);
	const pendingCount = $derived(pendingEdits.size);

	// Derived states for selection
	const allSelected = $derived(
		$fieldDeviceStore.items.length > 0 &&
			$fieldDeviceStore.items.every((d) => selectedIds.has(d.id))
	);
	const someSelected = $derived(
		$fieldDeviceStore.items.some((d) => selectedIds.has(d.id)) && !allSelected
	);
	const selectedCount = $derived(selectedIds.size);

	onMount(() => {
		fieldDeviceStore.load();
		// Preload lookup cache for better UX
		lookupCache.preloadAll();
	});

	function applyFilters() {
		fieldDeviceStore.setFilters({
			buildingId: buildingId || undefined,
			controlCabinetId: controlCabinetId || undefined,
			spsControllerId: spsControllerId || undefined,
			spsControllerSystemTypeId: spsControllerSystemTypeId || undefined
		});
	}

	function clearFilters() {
		buildingId = '';
		controlCabinetId = '';
		spsControllerId = '';
		spsControllerSystemTypeId = '';
		fieldDeviceStore.clearAllFilters();
	}

	function handleMultiCreateSuccess(createdDevices: FieldDevice[]) {
		showMultiCreateForm = false;
		fieldDeviceStore.reload();
		addToast(`Created ${createdDevices.length} field device(s).`, 'success');
	}

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

	function formatSPSControllerSystemType(device: FieldDevice): string {
		const sysType = device.sps_controller_system_type;
		if (!sysType) return '-';
		const number = sysType.number ?? '';
		const docName = sysType.document_name ?? '';
		if (number && docName) return `${number} - ${docName}`;
		if (number) return String(number);
		if (docName) return docName;
		return '-';
	}

	async function handleApparatChange(device: FieldDevice, newApparatId: string) {
		if (!newApparatId || newApparatId === device.apparat_id) return;
		try {
			await updateFieldDevice(device.id, {
				apparat_id: newApparatId
			});
			addToast('Apparat updated successfully', 'success');
			fieldDeviceStore.reload();
		} catch (error: any) {
			addToast(`Failed to update apparat: ${error.message}`, 'error');
		}
	}

	async function handleSystemPartChange(device: FieldDevice, newSystemPartId: string) {
		if (!newSystemPartId || newSystemPartId === device.system_part_id) return;
		try {
			await updateFieldDevice(device.id, {
				system_part_id: newSystemPartId
			});
			addToast('System Part updated successfully', 'success');
			fieldDeviceStore.reload();
		} catch (error: any) {
			addToast(`Failed to update system part: ${error.message}`, 'error');
		}
	}

	function handleSearchInput(e: Event) {
		const value = (e.target as HTMLInputElement).value;
		searchInput = value;
		fieldDeviceStore.search(value);
	}

	function handlePrevious() {
		if ($fieldDeviceStore.page > 1) {
			fieldDeviceStore.goToPage($fieldDeviceStore.page - 1);
		}
	}

	function handleNext() {
		if ($fieldDeviceStore.page < $fieldDeviceStore.totalPages) {
			fieldDeviceStore.goToPage($fieldDeviceStore.page + 1);
		}
	}

	// Reactive statement to check if any filters are active
	const hasActiveFilters = $derived(
		buildingId || controlCabinetId || spsControllerId || spsControllerSystemTypeId
	);

	// Selection functions
	function toggleSelectAll() {
		if (allSelected) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set($fieldDeviceStore.items.map((d) => d.id));
		}
	}

	function toggleSelect(id: string) {
		const newSet = new Set(selectedIds);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		selectedIds = newSet;
	}

	function clearSelection() {
		selectedIds = new Set();
	}

	// Column count for colspan (base columns + specification columns when shown)
	const baseColumnCount = 10;
	const specColumnCount = 11; // Additional columns for specification
	const columnCount = $derived(
		showSpecifications ? baseColumnCount + specColumnCount : baseColumnCount
	);

	// Inline editing functions
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

	// Specification edit helpers
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

	function isSpecFieldDirty(deviceId: string, field: keyof SpecificationInput): boolean {
		const edit = pendingEdits.get(deviceId);
		if (!edit) return false;
		const spec = (edit as Record<string, unknown>)._specification as SpecificationInput | undefined;
		return spec ? field in spec : false;
	}

	function getPendingSpecValue(
		deviceId: string,
		field: keyof SpecificationInput
	): string | undefined {
		const edit = pendingEdits.get(deviceId);
		if (!edit) return undefined;
		const spec = (edit as Record<string, unknown>)._specification as SpecificationInput | undefined;
		if (!spec || !(field in spec)) return undefined;
		const val = spec[field];
		return val !== undefined ? String(val) : undefined;
	}

	function isFieldDirty(deviceId: string, field: keyof UpdateFieldDeviceRequest): boolean {
		const edit = pendingEdits.get(deviceId);
		return edit ? field in edit : false;
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

	function getError(deviceId: string): string | undefined {
		return editErrors.get(deviceId);
	}

	async function saveAllPendingEdits() {
		if (pendingEdits.size === 0) return;

		const updates: BulkUpdateFieldDeviceItem[] = [];
		for (const [id, changes] of pendingEdits) {
			const spec = (changes as Record<string, unknown>)._specification as
				| SpecificationInput
				| undefined;
			updates.push({
				id,
				bmk: changes.bmk,
				description: changes.description,
				apparat_nr: changes.apparat_nr,
				specification: spec
			});
		}

		try {
			const result = await bulkUpdateFieldDevices({ updates });

			// Process results and track errors
			const newErrors = new Map<string, string>();
			const successIds = new Set<string>();

			for (const r of result.results) {
				if (r.success) {
					successIds.add(r.id);
				} else if (r.error) {
					newErrors.set(r.id, r.error);
				}
			}

			// Remove successful edits from pending, keep failed ones
			const remainingEdits = new Map(pendingEdits);
			for (const id of successIds) {
				remainingEdits.delete(id);
			}
			pendingEdits = remainingEdits;
			editErrors = newErrors;

			if (result.success_count > 0) {
				addToast(`Updated ${result.success_count} field device(s)`, 'success');
				fieldDeviceStore.reload();
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
		editErrors = new Map();
	}

	async function handleBulkDelete() {
		if (selectedIds.size === 0) return;
		if (!confirm(`Delete ${selectedIds.size} field device(s)? This action cannot be undone.`))
			return;

		try {
			const result = await bulkDeleteFieldDevices([...selectedIds]);
			if (result.success_count > 0) {
				addToast(`Deleted ${result.success_count} field device(s)`, 'success');
			}
			if (result.failure_count > 0) {
				addToast(`Failed to delete ${result.failure_count} device(s)`, 'error');
			}
			selectedIds = new Set();
			fieldDeviceStore.reload();
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Bulk delete failed: ${err.message}`, 'error');
		}
	}
</script>

<svelte:head>
	<title>Field Devices | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Field Devices</h1>
			<p class="text-sm text-muted-foreground">
				Manage field devices, BMK identifiers, and specifications.
			</p>
		</div>
		<div class="flex gap-2">
			{#if !showMultiCreateForm}
				<Button variant="outline" onclick={() => (showMultiCreateForm = true)}>
					<ListPlus class="mr-2 size-4" />
					Multi-Create
				</Button>
				<Button>
					<Plus class="mr-2 size-4" />
					New Field Device
				</Button>
			{/if}
		</div>
	</div>

	<!-- Multi-Create Form -->
	{#if showMultiCreateForm}
		<Card.Root>
			<Card.Header>
				<Card.Title>Multi-Create Field Devices</Card.Title>
				<Card.Description>
					Create multiple field devices at once with automatic apparat number assignment.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<FieldDeviceMultiCreateForm
					onSuccess={handleMultiCreateSuccess}
					onCancel={() => (showMultiCreateForm = false)}
				/>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Filter Card -->
	<Card.Root>
		<Card.Header>
			<Card.Title>Filters</Card.Title>
			<Card.Description>
				Filter field devices by building, control cabinet, SPS controller, or system type.
			</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
				<div class="flex flex-col gap-2">
					<label for="building-filter" class="text-sm font-medium">Building</label>
					<BuildingSelect bind:value={buildingId} width="w-full" />
				</div>
				<div class="flex flex-col gap-2">
					<label for="control-cabinet-filter" class="text-sm font-medium">Control Cabinet</label>
					<ControlCabinetSelect bind:value={controlCabinetId} width="w-full" />
				</div>
				<div class="flex flex-col gap-2">
					<label for="sps-controller-filter" class="text-sm font-medium">SPS Controller</label>
					<SPSControllerSelect bind:value={spsControllerId} width="w-full" />
				</div>
				<div class="flex flex-col gap-2">
					<label for="sps-controller-system-type-filter" class="text-sm font-medium">
						SPS Controller System Type
					</label>
					<SPSControllerSystemTypeSelect bind:value={spsControllerSystemTypeId} width="w-full" />
				</div>
			</div>
			<div class="mt-4 flex gap-2">
				<Button onclick={applyFilters}>Apply Filters</Button>
				{#if hasActiveFilters}
					<Button variant="outline" onclick={clearFilters}>
						<X class="mr-2 size-4" />
						Clear Filters
					</Button>
				{/if}
			</div>
		</Card.Content>
	</Card.Root>

	<!-- Data Table with Expandable Rows and Selection -->
	<div class="flex flex-col gap-4">
		<!-- Search Bar and Selection Actions -->
		<div class="flex items-center gap-4">
			<div class="relative flex-1">
				<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					type="search"
					placeholder="Search field devices..."
					class="pl-9"
					value={searchInput}
					oninput={handleSearchInput}
				/>
			</div>

			{#if selectedCount > 0}
				<div class="flex items-center gap-2">
					<span class="text-sm text-muted-foreground">{selectedCount} selected</span>
					<Button variant="outline" size="sm" onclick={clearSelection}>
						<X class="mr-1 h-4 w-4" />
						Clear
					</Button>
					<Button variant="destructive" size="sm" onclick={handleBulkDelete}>
						<Trash2 class="mr-1 h-4 w-4" />
						Delete
					</Button>
				</div>
			{/if}

			<Button
				variant="outline"
				onclick={() => fieldDeviceStore.reload()}
				disabled={$fieldDeviceStore.loading}
			>
				Refresh
			</Button>
		</div>

		<!-- Error Message -->
		{#if $fieldDeviceStore.error}
			<div
				class="rounded-md border border-destructive/50 bg-destructive/15 px-4 py-3 text-destructive"
			>
				<p class="font-medium">Error</p>
				<p class="text-sm">{$fieldDeviceStore.error}</p>
			</div>
		{/if}

		<!-- Table -->
		<div class="rounded-lg border bg-background">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<!-- Selection Checkbox -->
						<Table.Head class="w-10">
							<Checkbox
								checked={allSelected}
								indeterminate={someSelected}
								onCheckedChange={toggleSelectAll}
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
					{#if $fieldDeviceStore.loading && $fieldDeviceStore.items.length === 0}
						{#each Array(5) as _}
							<Table.Row>
								{#each Array(columnCount) as _}
									<Table.Cell>
										<Skeleton class="h-8 w-full" />
									</Table.Cell>
								{/each}
							</Table.Row>
						{/each}
					{:else if $fieldDeviceStore.items.length === 0}
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
						{#each $fieldDeviceStore.items as device (device.id)}
							<!-- Main Row -->
							<Table.Row
								class={[
									$fieldDeviceStore.loading ? 'opacity-60' : '',
									selectedIds.has(device.id) ? 'bg-muted/50' : ''
								]
									.filter(Boolean)
									.join(' ')}
							>
								<!-- Selection Checkbox -->
								<Table.Cell class="p-2">
									<Checkbox
										checked={selectedIds.has(device.id)}
										onCheckedChange={() => toggleSelect(device.id)}
										aria-label={`Select ${device.bmk || device.id}`}
									/>
								</Table.Cell>
								<!-- Expand button for BACnet Objects -->
								<Table.Cell class="p-2">
									<Button
										variant="ghost"
										size="sm"
										class="h-6 w-6 p-0"
										onclick={() => toggleRowExpansion(device.id)}
										title="Expand BACnet objects"
									>
										{#if expandedBacnetRows.has(device.id)}
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
										pendingValue={getPendingValue(device.id, 'bmk')}
										type="text"
										maxlength={10}
										isDirty={isFieldDirty(device.id, 'bmk')}
										onSave={(v) => queueEdit(device.id, 'bmk', v || undefined)}
									/>
								</Table.Cell>
								<!-- Description -->
								<Table.Cell class="max-w-48 p-1">
									<EditableCell
										value={device.description ?? ''}
										pendingValue={getPendingValue(device.id, 'description')}
										type="text"
										maxlength={250}
										isDirty={isFieldDirty(device.id, 'description')}
										onSave={(v) => queueEdit(device.id, 'description', v || undefined)}
									/>
								</Table.Cell>
								<!-- Apparat Nr -->
								<Table.Cell class="p-1">
									<EditableCell
										value={device.apparat_nr}
										pendingValue={getPendingValue(device.id, 'apparat_nr')}
										type="number"
										min={1}
										max={99}
										isDirty={isFieldDirty(device.id, 'apparat_nr')}
										error={getError(device.id)}
										onSave={(v) => queueEdit(device.id, 'apparat_nr', v ? parseInt(v) : undefined)}
									/>
								</Table.Cell>
								<!-- Apparat (combobox with cache) -->
								<Table.Cell>
									<CachedApparatSelect
										value={device.apparat_id}
										width="w-full"
										onValueChange={(newVal) => handleApparatChange(device, newVal)}
									/>
								</Table.Cell>
								<!-- System Part (combobox with cache) -->
								<Table.Cell>
									<CachedSystemPartSelect
										value={device.system_part_id || ''}
										width="w-full"
										onValueChange={(newVal) => handleSystemPartChange(device, newVal)}
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
								<!-- Specification columns (shown when toggled) - now editable -->
								{#if showSpecifications}
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.specification_supplier || ''}
											pendingValue={getPendingSpecValue(device.id, 'specification_supplier')}
											isDirty={isSpecFieldDirty(device.id, 'specification_supplier')}
											error={getError(device.id)}
											maxlength={250}
											onSave={(v) =>
												queueSpecEdit(device.id, 'specification_supplier', v || undefined)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.specification_brand || ''}
											pendingValue={getPendingSpecValue(device.id, 'specification_brand')}
											isDirty={isSpecFieldDirty(device.id, 'specification_brand')}
											error={getError(device.id)}
											maxlength={250}
											onSave={(v) =>
												queueSpecEdit(device.id, 'specification_brand', v || undefined)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.specification_type || ''}
											pendingValue={getPendingSpecValue(device.id, 'specification_type')}
											isDirty={isSpecFieldDirty(device.id, 'specification_type')}
											error={getError(device.id)}
											maxlength={250}
											onSave={(v) => queueSpecEdit(device.id, 'specification_type', v || undefined)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.additional_info_motor_valve || ''}
											pendingValue={getPendingSpecValue(device.id, 'additional_info_motor_valve')}
											isDirty={isSpecFieldDirty(device.id, 'additional_info_motor_valve')}
											error={getError(device.id)}
											maxlength={250}
											onSave={(v) =>
												queueSpecEdit(device.id, 'additional_info_motor_valve', v || undefined)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.additional_info_size?.toString() || ''}
											pendingValue={getPendingSpecValue(device.id, 'additional_info_size')}
											isDirty={isSpecFieldDirty(device.id, 'additional_info_size')}
											error={getError(device.id)}
											type="number"
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'additional_info_size',
													v ? parseInt(v) : undefined
												)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.additional_information_installation_location ||
												''}
											pendingValue={getPendingSpecValue(
												device.id,
												'additional_information_installation_location'
											)}
											isDirty={isSpecFieldDirty(
												device.id,
												'additional_information_installation_location'
											)}
											error={getError(device.id)}
											maxlength={250}
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'additional_information_installation_location',
													v || undefined
												)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.electrical_connection_ph?.toString() || ''}
											pendingValue={getPendingSpecValue(device.id, 'electrical_connection_ph')}
											isDirty={isSpecFieldDirty(device.id, 'electrical_connection_ph')}
											error={getError(device.id)}
											type="number"
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'electrical_connection_ph',
													v ? parseInt(v) : undefined
												)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.electrical_connection_acdc || ''}
											pendingValue={getPendingSpecValue(device.id, 'electrical_connection_acdc')}
											isDirty={isSpecFieldDirty(device.id, 'electrical_connection_acdc')}
											error={getError(device.id)}
											maxlength={2}
											placeholder="AC/DC"
											onSave={(v) =>
												queueSpecEdit(device.id, 'electrical_connection_acdc', v || undefined)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.electrical_connection_amperage?.toString() || ''}
											pendingValue={getPendingSpecValue(
												device.id,
												'electrical_connection_amperage'
											)}
											isDirty={isSpecFieldDirty(device.id, 'electrical_connection_amperage')}
											error={getError(device.id)}
											type="number"
											placeholder="A"
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'electrical_connection_amperage',
													v ? parseFloat(v) : undefined
												)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.electrical_connection_power?.toString() || ''}
											pendingValue={getPendingSpecValue(device.id, 'electrical_connection_power')}
											isDirty={isSpecFieldDirty(device.id, 'electrical_connection_power')}
											error={getError(device.id)}
											type="number"
											placeholder="W"
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'electrical_connection_power',
													v ? parseFloat(v) : undefined
												)}
										/>
									</Table.Cell>
									<Table.Cell class="text-xs">
										<EditableCell
											value={device.specification?.electrical_connection_rotation?.toString() || ''}
											pendingValue={getPendingSpecValue(
												device.id,
												'electrical_connection_rotation'
											)}
											isDirty={isSpecFieldDirty(device.id, 'electrical_connection_rotation')}
											error={getError(device.id)}
											type="number"
											placeholder="RPM"
											onSave={(v) =>
												queueSpecEdit(
													device.id,
													'electrical_connection_rotation',
													v ? parseInt(v) : undefined
												)}
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

							<!-- Expanded BACnet Objects Row -->
							{#if expandedBacnetRows.has(device.id)}
								<Table.Row
									class="bg-purple-50/50 hover:bg-purple-50/70 dark:bg-purple-950/20 dark:hover:bg-purple-950/30"
								>
									<Table.Cell colspan={columnCount} class="p-0">
										<div class="border-l-4 border-l-purple-500 py-4 pr-4 pl-14">
											<!-- BACnet Objects Section -->
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
											{#if device.bacnet_objects && device.bacnet_objects.length > 0}
												<div class="overflow-x-auto">
													<table class="w-full text-sm">
														<thead>
															<tr class="border-b text-left text-xs text-muted-foreground">
																<th class="pr-4 pb-2">Text Fix</th>
																<th class="pr-4 pb-2">Description</th>
																<th class="pr-4 pb-2">Software Type</th>
																<th class="pr-4 pb-2">Software Nr</th>
																<th class="pr-4 pb-2">Hardware Type</th>
																<th class="pr-4 pb-2">Hardware Qty</th>
																<th class="pr-4 pb-2">GMS Visible</th>
																<th class="pb-2">Optional</th>
															</tr>
														</thead>
														<tbody>
															{#each device.bacnet_objects as obj (obj.id)}
																<tr
																	class="border-b border-purple-100 last:border-0 dark:border-purple-900"
																>
																	<td class="py-2 pr-4 font-medium">{obj.text_fix}</td>
																	<td class="py-2 pr-4">{obj.description || '-'}</td>
																	<td class="py-2 pr-4">{obj.software_type}</td>
																	<td class="py-2 pr-4">{obj.software_number}</td>
																	<td class="py-2 pr-4">{obj.hardware_type}</td>
																	<td class="py-2 pr-4">{obj.hardware_quantity}</td>
																	<td class="py-2 pr-4">
																		{#if obj.gms_visible}
																			<span class="text-green-600">Yes</span>
																		{:else}
																			<span class="text-muted-foreground">No</span>
																		{/if}
																	</td>
																	<td class="py-2">
																		{#if obj.optional}
																			<span class="text-amber-600">Yes</span>
																		{:else}
																			<span class="text-muted-foreground">No</span>
																		{/if}
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
										</div>
									</Table.Cell>
								</Table.Row>
							{/if}
						{/each}
					{/if}
				</Table.Body>
			</Table.Root>
		</div>

		<!-- Pagination -->
		{#if $fieldDeviceStore.totalPages > 1}
			<div class="flex items-center justify-between">
				<div class="text-sm text-muted-foreground">
					Page {$fieldDeviceStore.page} of {$fieldDeviceStore.totalPages} • {$fieldDeviceStore.total}
					{$fieldDeviceStore.total === 1 ? 'item' : 'items'} total
				</div>
				<div class="flex items-center gap-2">
					<Button
						variant="outline"
						size="sm"
						disabled={$fieldDeviceStore.page <= 1 || $fieldDeviceStore.loading}
						onclick={handlePrevious}
					>
						<ChevronLeft class="mr-1 h-4 w-4" />
						Previous
					</Button>
					<Button
						variant="outline"
						size="sm"
						disabled={$fieldDeviceStore.page >= $fieldDeviceStore.totalPages ||
							$fieldDeviceStore.loading}
						onclick={handleNext}
					>
						Next
						<ChevronRight class="ml-1 h-4 w-4" />
					</Button>
				</div>
			</div>
		{/if}
	</div>

	<!-- Floating Save Bar -->
	{#if hasUnsavedChanges}
		<div
			class="fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 items-center gap-3 rounded-lg border bg-card px-4 py-3 shadow-lg"
		>
			<span class="text-sm font-medium"
				>{pendingCount} unsaved change{pendingCount !== 1 ? 's' : ''}</span
			>
			<Button size="sm" onclick={saveAllPendingEdits}>
				<Save class="mr-1 h-4 w-4" />
				Save All
			</Button>
			<Button variant="ghost" size="sm" onclick={discardAllEdits}>
				<Undo class="mr-1 h-4 w-4" />
				Discard
			</Button>
		</div>
	{/if}
</div>
