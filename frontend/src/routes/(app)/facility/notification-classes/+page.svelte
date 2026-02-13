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
	import { notificationClassesStore } from '$lib/stores/list/entityStores.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';
	import NotificationClassForm from '$lib/components/facility/NotificationClassForm.svelte';
	import { deleteNotificationClass } from '$lib/infrastructure/api/facility.adapter.js';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let showForm = $state(false);
	let editingItem: NotificationClass | undefined = $state(undefined);

	function handleEdit(item: NotificationClass) {
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
		notificationClassesStore.reload();
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

	async function handleDelete(item: NotificationClass) {
		const ok = await confirm({
			title: $t('facility.delete_notification_class_confirm').replace('{name}', ''),
			message: $t('facility.delete_notification_class_confirm').replace('{name}', item.event_category || ''),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteNotificationClass(item.id);
			addToast($t('facility.notification_class_deleted'), 'success');
			notificationClassesStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('facility.delete_notification_class_failed'), 'error');
		}
	}

	onMount(() => {
		notificationClassesStore.load();
	});
</script>

<svelte:head>
	<title>{$t('facility.notification_classes')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.notification_classes_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.notification_classes_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_notification_class')}
			</Button>
		{/if}
	</div>

	{#if showForm}
		<NotificationClassForm
			initialData={editingItem}
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$notificationClassesStore}
		columns={[
			{ key: 'event_category', label: $t('facility.event_category') },
			{ key: 'nc', label: $t('facility.nc') },
			{ key: 'object_description', label: $t('facility.object_description') },
			{ key: 'meaning', label: $t('facility.meaning') },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_notification_classes')}
		emptyMessage={$t('facility.no_notification_classes_found')}
		onSearch={(text) => notificationClassesStore.search(text)}
		onPageChange={(page) => notificationClassesStore.goToPage(page)}
		onReload={() => notificationClassesStore.reload()}
	>
		{#snippet rowSnippet(item: NotificationClass)}
			<Table.Cell class="font-medium">{item.event_category}</Table.Cell>
			<Table.Cell>{item.nc}</Table.Cell>
			<Table.Cell>{item.object_description}</Table.Cell>
			<Table.Cell>{item.meaning}</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(item.event_category ?? item.id)}>
						{$t('facility.copy')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => goto(`/facility/notification-classes/${item.id}`)}>
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
