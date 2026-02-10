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
	import { stateTextsStore } from '$lib/stores/list/entityStores.js';
	import type { StateText } from '$lib/domain/facility/index.js';
	import StateTextForm from '$lib/components/facility/StateTextForm.svelte';
	import { deleteStateText } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: StateText | undefined = $state(undefined);

	function handleEdit(item: StateText) {
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
		stateTextsStore.reload();
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

	async function handleDelete(item: StateText) {
		const ok = await confirm({
			title: 'Delete state text',
			message: `Delete ${item.ref_number}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteStateText(item.id);
			addToast('State text deleted', 'success');
			stateTextsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete state text', 'error');
		}
	}

	onMount(() => {
		stateTextsStore.load();
	});
</script>

<svelte:head>
	<title>State Texts | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">State Texts</h1>
			<p class="text-sm text-muted-foreground">Manage state text definitions and references.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New State Text
			</Button>
		{/if}
	</div>

	{#if showForm}
		<StateTextForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$stateTextsStore}
		columns={[
			{ key: 'ref_number', label: 'Ref Number' },
			{ key: 'state_text1', label: 'State Text' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
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
						<DropdownMenu.Item onclick={() => handleCopy(String(item.ref_number ?? item.id))}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/state-texts/${item.id}`)}>
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
