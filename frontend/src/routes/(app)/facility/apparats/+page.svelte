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
	import { apparatsStore } from '$lib/stores/list/entityStores.js';
	import type { Apparat } from '$lib/domain/facility/index.js';
	import ApparatForm from '$lib/components/facility/ApparatForm.svelte';
	import { deleteApparat } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: Apparat | undefined = $state(undefined);

	function handleEdit(item: Apparat) {
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
		apparatsStore.reload();
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

	async function handleDelete(item: Apparat) {
		const ok = await confirm({
			title: 'Delete apparat',
			message: `Delete ${item.short_name ?? item.name}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteApparat(item.id);
			addToast('Apparat deleted', 'success');
			apparatsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete apparat', 'error');
		}
	}

	onMount(() => {
		apparatsStore.load();
	});
</script>

<svelte:head>
	<title>Apparats | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Apparats</h1>
			<p class="text-sm text-muted-foreground">Manage apparats and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Apparat
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ApparatForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$apparatsStore}
		columns={[
			{ key: 'short_name', label: 'Short Name' },
			{ key: 'name', label: 'Name' },
			{ key: 'description', label: 'Description' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
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
						<DropdownMenu.Item onclick={() => handleCopy(item.short_name ?? item.id)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/apparats/${item.id}`)}>
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
