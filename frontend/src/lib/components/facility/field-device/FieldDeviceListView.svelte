<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { FileSpreadsheet, ListPlus } from '@lucide/svelte';
	import { createFieldDeviceStore } from '$lib/stores/facility/fieldDeviceStore.js';
	import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import { useUnsavedChangesWarning } from '$lib/hooks/useUnsavedChangesWarning.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';
	import type { FieldDeviceFilters } from '$lib/stores/facility/fieldDeviceStore.js';
	import FieldDeviceMultiCreateForm from './FieldDeviceMultiCreateForm.svelte';
	import FieldDeviceFilterCard from './FieldDeviceFilterCard.svelte';
	import FieldDeviceSearchBar from './FieldDeviceSearchBar.svelte';
	import FieldDeviceBulkEditPanel from './FieldDeviceBulkEditPanel.svelte';
	import FieldDeviceTable from './FieldDeviceTable.svelte';
	import FieldDevicePagination from './FieldDevicePagination.svelte';
	import FieldDeviceFloatingSaveBar from './FieldDeviceFloatingSaveBar.svelte';
	import FieldDeviceExportPanel from './FieldDeviceExportPanel.svelte';

	// Use Cases
	import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
	import { ListApparatsUseCase } from '$lib/application/useCases/facility/listApparatsUseCase.js';
	import { ListSystemPartsUseCase } from '$lib/application/useCases/facility/listSystemPartsUseCase.js';
	import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
	import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
	import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';

	const manageFieldDeviceUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
	const listApparatsUseCase = new ListApparatsUseCase(apparatRepository);
	const listSystemPartsUseCase = new ListSystemPartsUseCase(systemPartRepository);

	interface Props {
		projectId?: string;
	}

	const props: Props = $props();

	// Create a store instance scoped to this component (optionally project-scoped).
	const projectId = props.projectId;
	const store = createFieldDeviceStore(300, projectId);

	// Editing composable
	const editing = useFieldDeviceEditing(projectId);

	// Browser warning for unsaved changes
	useUnsavedChangesWarning(() => editing.hasUnsavedChanges);

	// Preloaded lookup data for table selects
	let allApparats = $state<Apparat[]>([]);
	let allSystemParts = $state<SystemPart[]>([]);

	// UI toggles
	let showMultiCreateForm = $state(false);
	let showBulkEditPanel = $state(false);
	let showExportPanel = $state(false);
	let searchInput = $state('');

	// Selection state
	let selectedIds = $state<Set<string>>(new Set());

	// Derived states for selection
	const allSelected = $derived(
		$store.items.length > 0 && $store.items.every((d) => selectedIds.has(d.id))
	);
	const someSelected = $derived($store.items.some((d) => selectedIds.has(d.id)) && !allSelected);
	const selectedCount = $derived(selectedIds.size);

	// Auto-hide bulk edit panel when nothing is selected
	$effect(() => {
		if (selectedCount === 0) {
			showBulkEditPanel = false;
		}
	});

	onMount(() => {
		store.load();
		// Load lookups
		listApparatsUseCase
			.execute({ pagination: { page: 1, pageSize: 1000 }, search: { text: '' } })
			.then((res) => (allApparats = res.items))
			.catch(console.error);
		listSystemPartsUseCase
			.execute({ pagination: { page: 1, pageSize: 1000 }, search: { text: '' } })
			.then((res) => (allSystemParts = res.items))
			.catch(console.error);
	});

	// Filter callbacks
	function handleApplyFilters(filters: FieldDeviceFilters) {
		store.setFilters(filters);
	}

	function handleClearFilters() {
		store.clearAllFilters();
	}

	// Multi-create
	async function handleMultiCreateSuccess(createdDevices: FieldDevice[]) {
		showMultiCreateForm = false;

		if (projectId) {
			try {
				await Promise.all(
					createdDevices.map((device) => projectRepository.addFieldDevice(projectId!, device.id))
				);
				addToast(
					`Created ${createdDevices.length} field device(s) and linked them to the project.`,
					'success'
				);
			} catch (err) {
				const message =
					err instanceof Error
						? err.message
						: 'Some field devices were created but could not be linked';
				addToast(`Failed to link field devices: ${message}`, 'error');
			}
		} else {
			addToast(`Created ${createdDevices.length} field device(s).`, 'success');
		}

		store.reload();
	}

	// Search
	function handleSearch(value: string) {
		searchInput = value;
		store.search(value);
	}

	// Sorting
	function handleSort(orderBy: string) {
		const currentOrderBy = $store.orderBy;
		const currentOrder = $store.order ?? 'asc';

		if (currentOrderBy !== orderBy) {
			store.setSort(orderBy, 'asc');
			return;
		}

		if (currentOrder === 'asc') {
			store.setSort(orderBy, 'desc');
			return;
		}

		store.setSort(undefined, undefined);
	}

	// Pagination
	function handlePrevious() {
		if ($store.page > 1) {
			store.goToPage($store.page - 1);
		}
	}

	function handleNext() {
		if ($store.page < $store.totalPages) {
			store.goToPage($store.page + 1);
		}
	}

	// Selection
	function toggleSelectAll() {
		if (allSelected) {
			selectedIds = new Set();
		} else {
			selectedIds = new Set($store.items.map((d) => d.id));
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

	// Bulk delete
	async function handleBulkDelete() {
		if (selectedIds.size === 0) return;
		if (!confirm(`Delete ${selectedIds.size} field device(s)? This action cannot be undone.`))
			return;

		try {
			const result = await manageFieldDeviceUseCase.bulkDelete([...selectedIds]);
			if (result.success_count > 0) {
				addToast(`Deleted ${result.success_count} field device(s)`, 'success');
			}
			if (result.failure_count > 0) {
				addToast(`Failed to delete ${result.failure_count} device(s)`, 'error');
			}
			selectedIds = new Set();
			store.reload();
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Bulk delete failed: ${err.message}`, 'error');
		}
	}

	async function handleCopy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}

	async function handleDelete(device: FieldDevice) {
		if (!confirm(`Delete ${device.bmk ?? device.id}? This action cannot be undone.`)) return;
		try {
			await manageFieldDeviceUseCase.delete(device.id);
			addToast('Field device deleted', 'success');
			const nextSelected = new Set(selectedIds);
			nextSelected.delete(device.id);
			selectedIds = nextSelected;
			store.reload();
		} catch (error: unknown) {
			const err = error as Error;
			addToast(`Delete failed: ${err.message}`, 'error');
		}
	}

	// Bulk edit toggle
	function toggleBulkEdit() {
		showBulkEditPanel = !showBulkEditPanel;
	}

	// Save / discard
	function handleSave() {
		editing.saveAllPendingEdits($store.items, (updated) => {
			updated.forEach((item) => store.updateItem(item));
		});
	}

	function handleDiscard() {
		editing.discardAllEdits();
	}
</script>

<div class="flex flex-col gap-6">
	<!-- Action Buttons -->
	<div class="flex justify-end gap-2">
		<Button variant="outline" onclick={() => (showExportPanel = !showExportPanel)}>
			<FileSpreadsheet class="mr-2 size-4" />
			{showExportPanel ? 'Hide Export' : 'Export Excel'}
		</Button>
		{#if !showMultiCreateForm}
			<Button variant="outline" onclick={() => (showMultiCreateForm = true)}>
				<ListPlus class="mr-2 size-4" />
				Multi-Create
			</Button>
		{/if}
	</div>

	{#if showExportPanel}
		<FieldDeviceExportPanel {projectId} />
	{/if}

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
					{projectId}
					onSuccess={handleMultiCreateSuccess}
					onCancel={() => (showMultiCreateForm = false)}
				/>
			</Card.Content>
		</Card.Root>
	{/if}

	<!-- Filter Card -->
	<FieldDeviceFilterCard
		onApplyFilters={handleApplyFilters}
		onClearFilters={handleClearFilters}
		showProjectFilter={!projectId}
	/>

	<!-- Data Table with Expandable Rows and Selection -->
	<div class="flex flex-col gap-4">
		<!-- Search Bar and Selection Actions -->
		<FieldDeviceSearchBar
			{searchInput}
			{selectedCount}
			loading={$store.loading}
			{showBulkEditPanel}
			onSearch={handleSearch}
			onClearSelection={clearSelection}
			onBulkDelete={handleBulkDelete}
			onToggleBulkEdit={toggleBulkEdit}
			onRefresh={() => store.reload()}
		/>

		<!-- Bulk Edit Panel -->
		{#if showBulkEditPanel && selectedCount > 0}
			<FieldDeviceBulkEditPanel
				{selectedCount}
				{selectedIds}
				{allApparats}
				{allSystemParts}
				{editing}
			/>
		{/if}

		<!-- Error Message -->
		{#if $store.error}
			<div
				class="rounded-md border border-destructive/50 bg-destructive/15 px-4 py-3 text-destructive"
			>
				<p class="font-medium">Error</p>
				<p class="text-sm">{$store.error}</p>
			</div>
		{/if}

		<!-- Table -->
		<FieldDeviceTable
			items={$store.items}
			loading={$store.loading}
			sortBy={$store.orderBy}
			sortOrder={$store.order}
			{searchInput}
			{allApparats}
			{allSystemParts}
			{selectedIds}
			{editing}
			{allSelected}
			{someSelected}
			onCopy={handleCopy}
			onDelete={handleDelete}
			onToggleSelect={toggleSelect}
			onToggleSelectAll={toggleSelectAll}
			onSort={handleSort}
		/>

		<!-- Pagination -->
		<FieldDevicePagination
			page={$store.page}
			totalPages={$store.totalPages}
			total={$store.total}
			loading={$store.loading}
			onPrevious={handlePrevious}
			onNext={handleNext}
		/>
	</div>

	<!-- Floating Save Bar -->
	<FieldDeviceFloatingSaveBar {editing} onSave={handleSave} onDiscard={handleDiscard} />
</div>
