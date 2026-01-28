<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { objectDataStore } from '$lib/stores/list/entityStores.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';
	import ObjectDataForm from '$lib/components/facility/ObjectDataForm.svelte';

	let showForm = $state(false);
	let editingItem: ObjectData | undefined = $state(undefined);

	function handleEdit(item: ObjectData) {
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
		objectDataStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		objectDataStore.load();
	});
</script>

<svelte:head>
	<title>Object Data | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Object Data</h1>
			<p class="text-sm text-muted-foreground">
				Manage object data configurations and BACnet objects.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Object Data
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ObjectDataForm initialData={editingItem} on:success={handleSuccess} on:cancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$objectDataStore}
		columns={[
			{ key: 'description', label: 'Description' },
			{ key: 'version', label: 'Version' },
			{ key: 'is_active', label: 'Status' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search object data..."
		emptyMessage="No object data found. Create your first object data to get started."
		onSearch={(text) => objectDataStore.search(text)}
		onPageChange={(page) => objectDataStore.goToPage(page)}
		onReload={() => objectDataStore.reload()}
	>
		{#snippet rowSnippet(item: ObjectData)}
			<Table.Cell class="font-medium">{item.description}</Table.Cell>
			<Table.Cell>
				<code class="rounded bg-muted px-1.5 py-0.5 text-sm">{item.version}</code>
			</Table.Cell>
			<Table.Cell>
				<span
					class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {item.is_active
						? 'bg-green-50 text-green-700'
						: 'bg-gray-50 text-gray-700'}"
				>
					{item.is_active ? 'Active' : 'Inactive'}
				</span>
			</Table.Cell>
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
