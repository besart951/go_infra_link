<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { specificationsStore } from '$lib/stores/list/entityStores.js';
	import type { Specification } from '$lib/domain/facility/index.js';

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
		<Button>
			<Plus class="mr-2 size-4" />
			New Specification
		</Button>
	</div>

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
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
