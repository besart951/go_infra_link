<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Plus, ListPlus } from '@lucide/svelte';
	import { createFieldDeviceStore } from '$lib/stores/facility/fieldDeviceStore.js';
	import {
		bulkDeleteFieldDevices,
		deleteFieldDevice,
		listApparats,
		listSystemParts
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { addProjectFieldDevice } from '$lib/infrastructure/api/project.adapter.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import { useUnsavedChangesWarning } from '$lib/hooks/useUnsavedChangesWarning.svelte.js';
	import type { FieldDevice, Apparat, SystemPart } from '$lib/domain/facility/index.js';
	import type { FieldDeviceFilters } from '$lib/stores/facility/fieldDeviceStore.js';
	import FieldDeviceMultiCreateForm from '$lib/components/facility/FieldDeviceMultiCreateForm.svelte';
	import FieldDeviceFilterCard from '$lib/components/facility/FieldDeviceFilterCard.svelte';
	import FieldDeviceSearchBar from '$lib/components/facility/FieldDeviceSearchBar.svelte';
	import FieldDeviceBulkEditPanel from '$lib/components/facility/FieldDeviceBulkEditPanel.svelte';
	import FieldDeviceTable from '$lib/components/facility/FieldDeviceTable.svelte';
	import FieldDevicePagination from '$lib/components/facility/FieldDevicePagination.svelte';
	import FieldDeviceFloatingSaveBar from '$lib/components/facility/FieldDeviceFloatingSaveBar.svelte';

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
		listApparats({ limit: 1000 }).then((res) => (allApparats = res.items));
		listSystemParts({ limit: 1000 }).then((res) => (allSystemParts = res.items));
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
					createdDevices.map((device) => addProjectFieldDevice(projectId!, device.id))
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
			const result = await bulkDeleteFieldDevices([...selectedIds]);
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
			await deleteFieldDevice(device.id);
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
