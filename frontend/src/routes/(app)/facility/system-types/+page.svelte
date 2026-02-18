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
	import SystemTypeForm from '$lib/components/facility/forms/SystemTypeForm.svelte';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
	const manageSystemType = new ManageEntityUseCase(systemTypeRepository);
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let showForm = $state(false);
	let editingItem: SystemType | undefined = $state(undefined);

	function formatNumber(value: number) {
		return String(value).padStart(4, '0');
	}

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
			title: $t('common.delete'),
			message: $t('facility.delete_system_type_confirm').replace('{name}', item.name),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await manageSystemType.delete(item.id);
			addToast($t('facility.system_type_deleted'), 'success');
			systemTypesStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('facility.delete_system_type_failed'), 'error');
		}
	}

	onMount(() => {
		systemTypesStore.load();
	});
</script>

<svelte:head>
	<title>{$t('facility.system_types_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.system_types_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.system_types_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_system_type')}
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SystemTypeForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$systemTypesStore}
		columns={[
			{ key: 'name', label: $t('common.name') },
			{ key: 'number_min', label: 'Min Number' },
			{ key: 'number_max', label: 'Max Number' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_system_types')}
		emptyMessage={$t('facility.no_system_types_found')}
		onSearch={(text) => systemTypesStore.search(text)}
		onPageChange={(page) => systemTypesStore.goToPage(page)}
		onReload={() => systemTypesStore.reload()}
	>
		{#snippet rowSnippet(item: SystemType)}
			<Table.Cell class="font-medium">{item.name}</Table.Cell>
			<Table.Cell>{formatNumber(item.number_min)}</Table.Cell>
			<Table.Cell>{formatNumber(item.number_max)}</Table.Cell>
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
						{$t('facility.copy')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => goto(`/facility/system-types/${item.id}`)}>
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
