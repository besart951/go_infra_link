<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { systemTypesStore } from '$lib/stores/list/entityStores.js';
	import type { SystemType } from '$lib/domain/facility/index.js';

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
		<Button>
			<Plus class="mr-2 size-4" />
			New System Type
		</Button>
	</div>

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
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
