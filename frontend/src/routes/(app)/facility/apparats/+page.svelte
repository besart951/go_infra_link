<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { apparatsStore } from '$lib/stores/list/entityStores.js';
	import type { Apparat } from '$lib/domain/facility/index.js';

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
		<Button>
			<Plus class="mr-2 size-4" />
			New Apparat
		</Button>
	</div>

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
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
