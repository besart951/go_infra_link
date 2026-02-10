<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { alarmDefinitionsStore } from '$lib/stores/list/entityStores.js';
	import type { AlarmDefinition } from '$lib/domain/facility/index.js';
	import AlarmDefinitionForm from '$lib/components/facility/AlarmDefinitionForm.svelte';
	import { deleteAlarmDefinition } from '$lib/infrastructure/api/facility.adapter.js';

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

	async function handleCopy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}

	async function handleDelete(item: AlarmDefinition) {
		const ok = await confirm({
			title: 'Delete alarm definition',
			message: `Delete ${item.name}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteAlarmDefinition(item.id);
			addToast('Alarm definition deleted', 'success');
			alarmDefinitionsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete alarm definition', 'error');
		}
	}

	onMount(() => {
		alarmDefinitionsStore.load();
	});
</script>

<svelte:head>
	<title>Alarm Definitions | Infra Link</title>
</svelte:head>

<ConfirmDialog />

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
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$alarmDefinitionsStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'alarm_note', label: 'Note' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
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
			<Table.Cell class="text-right">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger>
						{#snippet child({ props })}
							<Button variant="ghost" size="icon" {...props}>
								<EllipsisIcon class="size-4" />
							</Button>
						{/snippet}
					</DropdownMenu.Trigger>
					<DropdownMenu.Content align="end" class="w-40">
						<DropdownMenu.Item onclick={() => handleCopy(item.name ?? item.id)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/alarm-definitions/${item.id}`)}>
							View
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(item)}>Edit</DropdownMenu.Item>
						<DropdownMenu.Separator />
						<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(item)}>
							Delete
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
