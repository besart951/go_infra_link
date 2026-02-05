<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { buildingsStore } from '$lib/stores/list/entityStores.js';
	import type { Building } from '$lib/domain/facility/index.js';
	import BuildingForm from '$lib/components/facility/BuildingForm.svelte';

	let showForm = $state(false);
	let editingBuilding: Building | undefined = $state(undefined);

	function handleEdit(building: Building) {
		editingBuilding = building;
		showForm = true;
	}

	function handleCreate() {
		editingBuilding = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingBuilding = undefined;
		buildingsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingBuilding = undefined;
	}

	onMount(() => {
		buildingsStore.load();
	});
</script>

<svelte:head>
	<title>Buildings | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Buildings</h1>
			<p class="text-sm text-muted-foreground">Manage building infrastructure and IWS codes.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Building
			</Button>
		{/if}
	</div>

	{#if showForm}
		<BuildingForm initialData={editingBuilding} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$buildingsStore}
		columns={[
			{ key: 'iws_code', label: 'IWS Code' },
			{ key: 'building_group', label: 'Building Group' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search buildings..."
		emptyMessage="No buildings found. Create your first building to get started."
		onSearch={(text) => buildingsStore.search(text)}
		onPageChange={(page) => buildingsStore.goToPage(page)}
		onReload={() => buildingsStore.reload()}
	>
		{#snippet rowSnippet(building: Building)}
			<Table.Cell class="font-medium">
				<a href="/facility/buildings/{building.id}" class="hover:underline">
					{building.iws_code}
				</a>
			</Table.Cell>
			<Table.Cell>{building.building_group}</Table.Cell>
			<Table.Cell>
				{new Date(building.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(building)}>
						<Pencil class="size-4" />
					</Button>
					<Button variant="ghost" size="sm" href="/facility/buildings/{building.id}">View</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
