<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { apparatsStore } from '$lib/stores/list/entityStores.js';
	import type { Apparat } from '$lib/domain/facility/index.js';
	import ApparatForm from '$lib/components/facility/ApparatForm.svelte';

	let showForm = $state(false);
	let editingItem: Apparat | undefined = $state(undefined);

	function handleEdit(item: Apparat) {
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
		apparatsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		apparatsStore.load();
	});
</script>

<svelte:head>
	<title>Apparats | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Apparats</h1>
			<p class="text-sm text-muted-foreground">Manage apparats and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Apparat
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ApparatForm initialData={editingItem} on:success={handleSuccess} on:cancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$apparatsStore}
		columns={[
			{ key: 'short_name', label: 'Short Name' },
			{ key: 'name', label: 'Name' },
			{ key: 'description', label: 'Description' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search apparats..."
		emptyMessage="No apparats found. Create your first apparat to get started."
		onSearch={(text) => apparatsStore.search(text)}
		onPageChange={(page) => apparatsStore.goToPage(page)}
		onReload={() => apparatsStore.reload()}
	>
		{#snippet rowSnippet(item: Apparat)}
			<Table.Cell class="font-medium">{item.short_name}</Table.Cell>
			<Table.Cell>{item.name}</Table.Cell>
			<Table.Cell>{item.description ?? 'N/A'}</Table.Cell>
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
