<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { stateTextsStore } from '$lib/stores/list/entityStores.js';
	import type { StateText } from '$lib/domain/facility/index.js';

	onMount(() => {
		stateTextsStore.load();
	});
</script>

<svelte:head>
	<title>State Texts | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">State Texts</h1>
			<p class="text-sm text-muted-foreground">Manage state text definitions and references.</p>
		</div>
		<Button>
			<Plus class="mr-2 size-4" />
			New State Text
		</Button>
	</div>

	<PaginatedList
		state={$stateTextsStore}
		columns={[
			{ key: 'ref_number', label: 'Ref Number' },
			{ key: 'state_text1', label: 'State Text' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search state texts..."
		emptyMessage="No state texts found. Create your first state text to get started."
		onSearch={(text) => stateTextsStore.search(text)}
		onPageChange={(page) => stateTextsStore.goToPage(page)}
		onReload={() => stateTextsStore.reload()}
	>
		{#snippet rowSnippet(item: StateText)}
			<Table.Cell class="font-medium">{item.ref_number}</Table.Cell>
			<Table.Cell>{item.state_text1 ?? 'N/A'}</Table.Cell>
			<Table.Cell>
				{new Date(item.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
