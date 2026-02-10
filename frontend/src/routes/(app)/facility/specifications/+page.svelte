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
	import { specificationsStore } from '$lib/stores/list/entityStores.js';
	import type { Specification } from '$lib/domain/facility/index.js';
	import SpecificationForm from '$lib/components/facility/SpecificationForm.svelte';
	import { deleteSpecification } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: Specification | undefined = $state(undefined);

	function handleEdit(item: Specification) {
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
		specificationsStore.reload();
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

	async function handleDelete(item: Specification) {
		const ok = await confirm({
			title: 'Delete specification',
			message: `Delete ${item.specification_supplier ?? item.specification_type ?? 'specification'}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteSpecification(item.id);
			addToast('Specification deleted', 'success');
			specificationsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete specification', 'error');
		}
	}

	onMount(() => {
		specificationsStore.load();
	});
</script>

<svelte:head>
	<title>Specifications | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Specifications</h1>
			<p class="text-sm text-muted-foreground">
				Manage technical specifications for field devices.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Specification
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SpecificationForm
			initialData={editingItem}
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$specificationsStore}
		columns={[
			{ key: 'supplier', label: 'Supplier' },
			{ key: 'brand', label: 'Brand' },
			{ key: 'type', label: 'Type' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search specifications..."
		emptyMessage="No specifications found. Create your first specification to get started."
		onSearch={(text) => specificationsStore.search(text)}
		onPageChange={(page) => specificationsStore.goToPage(page)}
		onReload={() => specificationsStore.reload()}
	>
		{#snippet rowSnippet(item: Specification)}
			<Table.Cell class="font-medium">{item.specification_supplier ?? 'N/A'}</Table.Cell>
			<Table.Cell>{item.specification_brand ?? 'N/A'}</Table.Cell>
			<Table.Cell>{item.specification_type ?? 'N/A'}</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(item.specification_supplier ?? item.id)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/specifications/${item.id}`)}>
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
