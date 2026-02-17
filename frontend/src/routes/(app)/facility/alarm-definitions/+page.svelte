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
	import { ManageAlarmDefinitionUseCase } from '$lib/application/useCases/facility/manageAlarmDefinitionUseCase.js';
	import { alarmDefinitionRepository } from '$lib/infrastructure/api/alarmDefinitionRepository.js';
	const manageAlarmDefinition = new ManageAlarmDefinitionUseCase(alarmDefinitionRepository);
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

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
			title: $t('facility.delete_alarm_definition_confirm').replace('{name}', ''),
			message: $t('facility.delete_alarm_definition_confirm').replace('{name}', item.name || ''),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await manageAlarmDefinition.delete(item.id);
			addToast($t('facility.alarm_definition_deleted'), 'success');
			alarmDefinitionsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('facility.delete_alarm_definition_failed'), 'error');
		}
	}

	onMount(() => {
		alarmDefinitionsStore.load();
	});
</script>

<svelte:head>
	<title>{$t('facility.alarm_definitions')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.alarm_definitions_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.alarm_definitions_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_alarm_definition')}
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
			{ key: 'name', label: $t('common.name') },
			{ key: 'alarm_note', label: $t('facility.alarm_note') },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_alarm_definitions')}
		emptyMessage={$t('facility.no_alarm_definitions_found')}
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
						{$t('facility.copy')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => goto(`/facility/alarm-definitions/${item.id}`)}>
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
