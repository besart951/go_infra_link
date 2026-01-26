<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { notificationClassesStore } from '$lib/stores/list/entityStores.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';

	onMount(() => {
		notificationClassesStore.load();
	});
</script>

<svelte:head>
	<title>Notification Classes | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Notification Classes</h1>
			<p class="text-sm text-muted-foreground">Manage notification classes and event categories.</p>
		</div>
		<Button>
			<Plus class="mr-2 size-4" />
			New Notification Class
		</Button>
	</div>

	<PaginatedList
		state={$notificationClassesStore}
		columns={[
			{ key: 'event_category', label: 'Event Category' },
			{ key: 'nc', label: 'NC' },
			{ key: 'object_description', label: 'Description' },
			{ key: 'meaning', label: 'Meaning' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search notification classes..."
		emptyMessage="No notification classes found. Create your first notification class to get started."
		onSearch={(text) => notificationClassesStore.search(text)}
		onPageChange={(page) => notificationClassesStore.goToPage(page)}
		onReload={() => notificationClassesStore.reload()}
	>
		{#snippet rowSnippet(item: NotificationClass)}
			<Table.Cell class="font-medium">{item.event_category}</Table.Cell>
			<Table.Cell>{item.nc}</Table.Cell>
			<Table.Cell>{item.object_description}</Table.Cell>
			<Table.Cell>{item.meaning}</Table.Cell>
			<Table.Cell>
				{new Date(item.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
