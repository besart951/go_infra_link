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
	import StateTextForm from '$lib/components/facility/forms/StateTextForm.svelte';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
	const manageStateText = new ManageEntityUseCase(stateTextRepository);
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

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
			title: $t('facility.delete_state_text_confirm').replace('{ref}', ''),
			message: $t('facility.delete_state_text_confirm').replace(
				'{ref}',
				String(item.ref_number || '')
			),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await manageStateText.delete(item.id);
			addToast($t('facility.state_text_deleted'), 'success');
			stateTextsStore.reload();
		} catch (err) {
			addToast(
				err instanceof Error ? err.message : $t('facility.delete_state_text_failed'),
				'error'
			);
		}
	}

	onMount(() => {
		stateTextsStore.load();
	});
</script>

<svelte:head>
	<title>{$t('facility.state_texts')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.state_texts_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.state_texts_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_state_text')}
			</Button>
		{/if}
	</div>

	{#if showForm}
		<StateTextForm initialData={editingItem} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$stateTextsStore}
		columns={[
			{ key: 'ref_number', label: $t('facility.ref_number') },
			{ key: 'state_text1', label: $t('facility.state_text1') },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_state_texts')}
		emptyMessage={$t('facility.no_state_texts_found')}
		onSearch={(text) => stateTextsStore.search(text)}
		onPageChange={(page) => stateTextsStore.goToPage(page)}
		onReload={() => stateTextsStore.reload()}
	>
		{#snippet rowSnippet(item: StateText)}
			<Table.Cell class="font-medium">{item.ref_number}</Table.Cell>
			<Table.Cell>{item.state_text1 ?? $t('common.not_available')}</Table.Cell>
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
							{$t('facility.copy')}
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/state-texts/${item.id}`)}>
							{$t('facility.view')}
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(item)}
							>{$t('common.edit')}</DropdownMenu.Item
						>
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
