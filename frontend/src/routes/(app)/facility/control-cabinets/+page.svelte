<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { controlCabinetsStore } from '$lib/stores/list/entityStores.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import ControlCabinetForm from '$lib/components/facility/ControlCabinetForm.svelte';

	let showForm = $state(false);
	let editingItem: ControlCabinet | undefined = $state(undefined);

	function handleEdit(item: ControlCabinet) {
		editingItem = item;
		showForm = true;
	}

	function handleCreate() {
		editingItem = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingItem = undefined;
		controlCabinetsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		controlCabinetsStore.load();
	});
</script>

<svelte:head>
	<title>Control Cabinets | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Control Cabinets</h1>
			<p class="text-sm text-muted-foreground">Manage control cabinets and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Control Cabinet
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ControlCabinetForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$controlCabinetsStore}
		columns={[
			{ key: 'cabinet_nr', label: 'Cabinet Nr' },
			{ key: 'building', label: 'Building ID' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search control cabinets..."
		emptyMessage="No control cabinets found. Create your first control cabinet to get started."
		onSearch={(text) => controlCabinetsStore.search(text)}
		onPageChange={(page) => controlCabinetsStore.goToPage(page)}
		onReload={() => controlCabinetsStore.reload()}
	>
		{#snippet rowSnippet(cabinet: ControlCabinet)}
			<Table.Cell class="font-medium">
				<a href="/facility/control-cabinets/{cabinet.id}" class="hover:underline">
					{cabinet.control_cabinet_nr ?? 'N/A'}
				</a>
			</Table.Cell>
			<Table.Cell>{cabinet.building_id}</Table.Cell>
			<Table.Cell>
				{new Date(cabinet.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(cabinet)}>
						<Pencil class="size-4" />
					</Button>
					<Button variant="ghost" size="sm" href="/facility/control-cabinets/{cabinet.id}">
						View
					</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
