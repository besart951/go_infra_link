<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from 'lucide-svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { spsControllersStore } from '$lib/stores/list/entityStores.js';
	import type { SPSController } from '$lib/domain/entities/spsController.js';
	import SPSControllerForm from '$lib/components/facility/SPSControllerForm.svelte';

	let showForm = $state(false);
	let editingItem: SPSController | undefined = $state(undefined);

	function handleEdit(item: SPSController) {
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
		spsControllersStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		spsControllersStore.load();
	});
</script>

<svelte:head>
	<title>SPS Controllers | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">SPS Controllers</h1>
			<p class="text-sm text-muted-foreground">
				Manage SPS controller devices and their configurations.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New SPS Controller
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SPSControllerForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$spsControllersStore}
		columns={[
			{ key: 'device_name', label: 'Device Name' },
			{ key: 'ga_device', label: 'GA Device' },
			{ key: 'ip_address', label: 'IP Address' },
			{ key: 'cabinet', label: 'Cabinet' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search SPS controllers..."
		emptyMessage="No SPS controllers found. Create your first SPS controller to get started."
		onSearch={(text) => spsControllersStore.search(text)}
		onPageChange={(page) => spsControllersStore.goToPage(page)}
		onReload={() => spsControllersStore.reload()}
	>
		{#snippet rowSnippet(controller: SPSController)}
			<Table.Cell class="font-medium">
				<a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
					{controller.device_name}
				</a>
			</Table.Cell>
			<Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
			<Table.Cell>
				{#if controller.ip_address}
					<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
						{controller.ip_address}
					</code>
				{:else}
					-
				{/if}
			</Table.Cell>
			<Table.Cell>{controller.control_cabinet_id}</Table.Cell>
			<Table.Cell>
				{new Date(controller.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(controller)}>
						<Pencil class="size-4" />
					</Button>
					<Button variant="ghost" size="sm" href="/facility/sps-controllers/{controller.id}">
						View
					</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
