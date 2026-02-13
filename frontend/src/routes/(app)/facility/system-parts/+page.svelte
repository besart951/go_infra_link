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
	import { systemPartsStore } from '$lib/stores/list/entityStores.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import SystemPartForm from '$lib/components/facility/SystemPartForm.svelte';
	import { deleteSystemPart } from '$lib/infrastructure/api/facility.adapter.js';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let showForm = $state(false);
	let editingItem: SystemPart | undefined = $state(undefined);

	function handleEdit(item: SystemPart) {
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
		systemPartsStore.reload();
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

	async function handleDelete(item: SystemPart) {
		const ok = await confirm({
			title: $t('common.delete'),
			message: $t('facility.delete_system_part_confirm').replace('{name}', item.short_name ?? item.name),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteSystemPart(item.id);
			addToast($t('facility.system_part_deleted'), 'success');
			systemPartsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('facility.delete_system_part_failed'), 'error');
		}
	}

	onMount(() => {
		systemPartsStore.load();
	});
</script>

<svelte:head>
	<title>{$t('facility.system_parts_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.system_parts_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.system_parts_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_system_part')}
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SystemPartForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$systemPartsStore}
		columns={[
			{ key: 'short_name', label: $t('facility.short_name') },
			{ key: 'name', label: $t('common.name') },
			{ key: 'description', label: $t('common.description') },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_system_parts')}
		emptyMessage={$t('facility.no_system_parts_found')}
		onSearch={(text) => systemPartsStore.search(text)}
		onPageChange={(page) => systemPartsStore.goToPage(page)}
		onReload={() => systemPartsStore.reload()}
	>
		{#snippet rowSnippet(item: SystemPart)}
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
						{$t('facility.copy')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => goto(`/facility/system-parts/${item.id}`)}>
						{$t('facility.view')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => handleEdit(item)}>{$t('common.edit')}</DropdownMenu.Item>
					<DropdownMenu.Separator />
					<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(item)}>
						{$t('common.delete')}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
