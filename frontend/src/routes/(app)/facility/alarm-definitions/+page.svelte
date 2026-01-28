<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { alarmDefinitionsStore } from '$lib/stores/list/entityStores.js';
	import type { AlarmDefinition } from '$lib/domain/facility/index.js';
	import AlarmDefinitionForm from '$lib/components/facility/AlarmDefinitionForm.svelte';

	let showForm = $state(false);
	let editingItem: AlarmDefinition | undefined = $state(undefined);

	function handleEdit(item: AlarmDefinition) {
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
		alarmDefinitionsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	onMount(() => {
		alarmDefinitionsStore.load();
	});
</script>

<svelte:head>
	<title>Alarm Definitions | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Alarm Definitions</h1>
			<p class="text-sm text-muted-foreground">Manage alarm definitions and notifications.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Alarm Definition
			</Button>
		{/if}
	</div>

	{#if showForm}
		<AlarmDefinitionForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$alarmDefinitionsStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'alarm_note', label: 'Note' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search alarm definitions..."
		emptyMessage="No alarm definitions found. Create your first alarm definition to get started."
		onSearch={(text) => alarmDefinitionsStore.search(text)}
		onPageChange={(page) => alarmDefinitionsStore.goToPage(page)}
		onReload={() => alarmDefinitionsStore.reload()}
	>
		{#snippet rowSnippet(item: AlarmDefinition)}
			<Table.Cell class="font-medium">{item.name}</Table.Cell>
			<Table.Cell>{item.alarm_note ?? 'N/A'}</Table.Cell>
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
