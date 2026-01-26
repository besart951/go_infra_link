<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { systemPartsStore } from '$lib/stores/list/entityStores.js';
	import type { SystemPart } from '$lib/domain/entities/systemPart.js';

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
		<Button>
			<Plus class="mr-2 size-4" />
			New System Part
		</Button>
	</div>

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
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
