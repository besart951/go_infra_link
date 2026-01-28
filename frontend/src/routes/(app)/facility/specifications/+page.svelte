<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { specificationsStore } from '$lib/stores/list/entityStores.js';
	import type { Specification } from '$lib/domain/facility/index.js';
	import SpecificationForm from '$lib/components/facility/SpecificationForm.svelte';

	let showForm = $state(false);
	let editingItem: Specification | undefined = $state(undefined);

	function handleEdit(item: Specification) {
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
		specificationsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		specificationsStore.load();
	});
</script>

<svelte:head>
	<title>Specifications | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Specifications</h1>
			<p class="text-sm text-muted-foreground">
				Manage technical specifications for field devices.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Specification
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SpecificationForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$specificationsStore}
		columns={[
			{ key: 'supplier', label: 'Supplier' },
			{ key: 'brand', label: 'Brand' },
			{ key: 'type', label: 'Type' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search specifications..."
		emptyMessage="No specifications found. Create your first specification to get started."
		onSearch={(text) => specificationsStore.search(text)}
		onPageChange={(page) => specificationsStore.goToPage(page)}
		onReload={() => specificationsStore.reload()}
	>
		{#snippet rowSnippet(item: Specification)}
			<Table.Cell class="font-medium">{item.specification_supplier ?? 'N/A'}</Table.Cell>
			<Table.Cell>{item.specification_brand ?? 'N/A'}</Table.Cell>
			<Table.Cell>{item.specification_type ?? 'N/A'}</Table.Cell>
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
