<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { fieldDevicesStore } from '$lib/stores/list/entityStores.js';
	import type { FieldDevice } from '$lib/domain/facility/index.js';

	onMount(() => {
		fieldDevicesStore.load();
	});
</script>

<svelte:head>
	<title>Field Devices | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Field Devices</h1>
			<p class="text-sm text-muted-foreground">
				Manage field devices, BMK identifiers, and specifications.
			</p>
		</div>
		<Button>
			<Plus class="mr-2 size-4" />
			New Field Device
		</Button>
	</div>

	<PaginatedList
		state={$fieldDevicesStore}
		columns={[
			{ key: 'bmk', label: 'BMK' },
			{ key: 'description', label: 'Description' },
			{ key: 'apparat_nr', label: 'Apparat Nr' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search field devices..."
		emptyMessage="No field devices found. Create your first field device to get started."
		onSearch={(text) => fieldDevicesStore.search(text)}
		onPageChange={(page) => fieldDevicesStore.goToPage(page)}
		onReload={() => fieldDevicesStore.reload()}
	>
		{#snippet rowSnippet(device: FieldDevice)}
			<Table.Cell class="font-medium">{device.bmk}</Table.Cell>
			<Table.Cell>{device.description}</Table.Cell>
			<Table.Cell>
				<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
					{device.apparat_nr}
				</code>
			</Table.Cell>
			<Table.Cell>
				{new Date(device.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<Button variant="ghost" size="sm">View</Button>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
