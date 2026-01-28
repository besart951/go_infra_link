<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { phaseListStore } from '$lib/stores/phases/phaseListStore.js';
	import type { Phase } from '$lib/domain/phase/index.js';

	onMount(() => {
		phaseListStore.load();
	});
</script>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold">Phases</h1>
		<Button onclick={() => goto('/projects/phases/new')}>Create Phase</Button>
	</div>

	<PaginatedList
		state={$phaseListStore}
		columns={[{ key: 'phase', label: 'Phase' }]}
		searchPlaceholder="Search phases..."
		emptyMessage="No phases found. Create your first phase to get started."
		onSearch={(text) => phaseListStore.search(text)}
		onPageChange={(page) => phaseListStore.goToPage(page)}
		onReload={() => phaseListStore.reload()}
	>
		{#snippet rowSnippet(phase: Phase)}
			<Table.Cell class="font-medium">{phase.id}</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
