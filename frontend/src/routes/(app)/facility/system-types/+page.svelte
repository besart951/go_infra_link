<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { systemTypesStore } from '$lib/stores/list/entityStores.js';
	import type { SystemType } from '$lib/domain/facility/index.js';
	import SystemTypeForm from '$lib/components/facility/SystemTypeForm.svelte';

	let showForm = $state(false);
	let editingItem: SystemType | undefined = $state(undefined);

	function handleEdit(item: SystemType) {
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
		systemTypesStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		systemTypesStore.load();
	});
</script>

<svelte:head>
	<title>System Types | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">System Types</h1>
			<p class="text-sm text-muted-foreground">Manage system types and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New System Type
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SystemTypeForm initialData={editingItem} on:success={handleSuccess} on:cancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$systemTypesStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'number_min', label: 'Min Number' },
			{ key: 'number_max', label: 'Max Number' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search system types..."
		emptyMessage="No system types found. Create your first system type to get started."
		onSearch={(text) => systemTypesStore.search(text)}
		onPageChange={(page) => systemTypesStore.goToPage(page)}
		onReload={() => systemTypesStore.reload()}
	>
		{#snippet rowSnippet(item: SystemType)}
			<Table.Cell class="font-medium">{item.name}</Table.Cell>
			<Table.Cell>{item.number_min}</Table.Cell>
			<Table.Cell>{item.number_max}</Table.Cell>
			<Table.Cell>
				{new Date(item.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(item)}>
						<Pencil class="size-4" />
					</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
