<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Plus, X, ListPlus } from 'lucide-svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { fieldDeviceStore } from '$lib/stores/facility/fieldDeviceStore.js';
	import { addToast } from '$lib/components/toast.svelte';
	import type { FieldDevice } from '$lib/domain/facility/index.js';
	import BuildingSelect from '$lib/components/facility/BuildingSelect.svelte';
	import ControlCabinetSelect from '$lib/components/facility/ControlCabinetSelect.svelte';
	import SPSControllerSelect from '$lib/components/facility/SPSControllerSelect.svelte';
	import SPSControllerSystemTypeSelect from '$lib/components/facility/SPSControllerSystemTypeSelect.svelte';
	import FieldDeviceMultiCreateForm from '$lib/components/facility/FieldDeviceMultiCreateForm.svelte';

	let buildingId = $state('');
	let controlCabinetId = $state('');
	let spsControllerId = $state('');
	let spsControllerSystemTypeId = $state('');
	let showMultiCreateDialog = $state(false);

	onMount(() => {
		fieldDeviceStore.load();
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
		showMultiCreateDialog = false;
		fieldDeviceStore.reload();
		addToast({
			type: 'success',
			message: 'Field devices created successfully',
			description: `Created ${createdDevices.length} field device(s)`
		});
	}

	// Reactive statement to check if any filters are active
	const hasActiveFilters = $derived(
		buildingId || controlCabinetId || spsControllerId || spsControllerSystemTypeId
	);
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
			<Button variant="outline" onclick={() => (showMultiCreateDialog = true)}>
				<ListPlus class="mr-2 size-4" />
				Multi-Create
			</Button>
			<Button>
				<Plus class="mr-2 size-4" />
				New Field Device
			</Button>
		</div>
	</div>

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

	<PaginatedList
		state={$fieldDeviceStore}
		columns={[
			{ key: 'bmk', label: 'BMK' },
			{ key: 'description', label: 'Description' },
			{ key: 'apparat_nr', label: 'Apparat Nr' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search field devices..."
		emptyMessage="No field devices found. Create your first field device to get started."
		onSearch={(text) => fieldDeviceStore.search(text)}
		onPageChange={(page) => fieldDeviceStore.goToPage(page)}
		onReload={() => fieldDeviceStore.reload()}
	>
		{#snippet rowSnippet(device: FieldDevice)}
			<Table.Cell class="font-medium">{device.bmk}</Table.Cell>
			<Table.Cell>{device.description}</Table.Cell>
			<Table.Cell>
				<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
					{device.apparat_nr}
				</code>
			</Table.Cell>
			<Table.Cell>
				{new Date(device.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>

<!-- Multi-Create Dialog -->
<Dialog.Root bind:open={showMultiCreateDialog}>
	<Dialog.Content class="max-h-[90vh] max-w-5xl overflow-y-auto">
		<Dialog.Header>
			<Dialog.Title>Multi-Create Field Devices</Dialog.Title>
			<Dialog.Description>
				Create multiple field devices at once with automatic apparat number assignment.
			</Dialog.Description>
		</Dialog.Header>
		<FieldDeviceMultiCreateForm
			onSuccess={handleMultiCreateSuccess}
			onCancel={() => (showMultiCreateDialog = false)}
		/>
	</Dialog.Content>
</Dialog.Root>
