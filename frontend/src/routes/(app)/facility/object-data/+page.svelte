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
	import { objectDataStore } from '$lib/stores/list/entityStores.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';
	import { deleteObjectData, getObjectData } from '$lib/infrastructure/api/facility.adapter.js';
	import ObjectDataForm from '$lib/components/facility/ObjectDataForm.svelte';

	let showForm = $state(false);
	let editingItem: ObjectData | undefined = $state(undefined);

	async function handleEdit(item: ObjectData) {
		try {
			editingItem = await getObjectData(item.id);
		} catch (error) {
			console.error(error);
			editingItem = item;
		}
		showForm = true;
	}

	function handleCreate() {
		editingItem = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingItem = undefined;
		objectDataStore.reload();
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

	async function handleDelete(item: ObjectData) {
		const ok = await confirm({
			title: 'Delete object data',
			message: `Delete ${item.description}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteObjectData(item.id);
			addToast('Object data deleted', 'success');
			objectDataStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete object data', 'error');
		}
	}

	onMount(() => {
		objectDataStore.load();
	});
</script>

<svelte:head>
	<title>Object Data | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Object Data</h1>
			<p class="text-sm text-muted-foreground">
				Manage object data configurations and BACnet objects.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Object Data
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ObjectDataForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$objectDataStore}
		columns={[
			{ key: 'description', label: 'Description' },
			{ key: 'version', label: 'Version' },
			{ key: 'is_active', label: 'Status' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search object data..."
		emptyMessage="No object data found. Create your first object data to get started."
		onSearch={(text) => objectDataStore.search(text)}
		onPageChange={(page) => objectDataStore.goToPage(page)}
		onReload={() => objectDataStore.reload()}
	>
		{#snippet rowSnippet(item: ObjectData)}
			<Table.Cell class="font-medium">{item.description}</Table.Cell>
			<Table.Cell>
				<code class="rounded bg-muted px-1.5 py-0.5 text-sm">{item.version}</code>
			</Table.Cell>
			<Table.Cell>
				<span
					class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {item.is_active
						? 'bg-green-50 text-green-700'
						: 'bg-gray-50 text-gray-700'}"
				>
					{item.is_active ? 'Active' : 'Inactive'}
				</span>
			</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(item.description ?? item.id)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/object-data/${item.id}`)}>
							View
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(item)}>
							Edit
						</DropdownMenu.Item>
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
