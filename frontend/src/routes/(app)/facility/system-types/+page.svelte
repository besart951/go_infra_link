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
	import { systemTypesStore } from '$lib/stores/list/entityStores.js';
	import type { SystemType } from '$lib/domain/facility/index.js';
	import SystemTypeForm from '$lib/components/facility/SystemTypeForm.svelte';
	import { deleteSystemType } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: SystemType | undefined = $state(undefined);

	function handleEdit(item: SystemType) {
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
		systemTypesStore.reload();
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

	async function handleDelete(item: SystemType) {
		const ok = await confirm({
			title: 'Delete system type',
			message: `Delete ${item.name}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteSystemType(item.id);
			addToast('System type deleted', 'success');
			systemTypesStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete system type', 'error');
		}
	}

	onMount(() => {
		systemTypesStore.load();
	});
</script>

<svelte:head>
	<title>System Types | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">System Types</h1>
			<p class="text-sm text-muted-foreground">Manage system types and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New System Type
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SystemTypeForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$systemTypesStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'number_min', label: 'Min Number' },
			{ key: 'number_max', label: 'Max Number' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
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
						<DropdownMenu.Item onclick={() => goto(`/facility/system-types/${item.id}`)}>
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
