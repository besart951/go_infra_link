<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { systemPartsStore } from '$lib/stores/list/entityStores.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import SystemPartForm from '$lib/components/facility/SystemPartForm.svelte';

	let showForm = $state(false);
	let editingItem: SystemPart | undefined = $state(undefined);

	function handleEdit(item: SystemPart) {
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
		systemPartsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		systemPartsStore.load();
	});
</script>

<svelte:head>
	<title>System Parts | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">System Parts</h1>
			<p class="text-sm text-muted-foreground">Manage system parts and components.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New System Part
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SystemPartForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$systemPartsStore}
		columns={[
			{ key: 'short_name', label: 'Short Name' },
			{ key: 'name', label: 'Name' },
			{ key: 'description', label: 'Description' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search system parts..."
		emptyMessage="No system parts found. Create your first system part to get started."
		onSearch={(text) => systemPartsStore.search(text)}
		onPageChange={(page) => systemPartsStore.goToPage(page)}
		onReload={() => systemPartsStore.reload()}
	>
		{#snippet rowSnippet(item: SystemPart)}
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
