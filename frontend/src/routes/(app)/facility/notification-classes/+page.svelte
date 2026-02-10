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
			title: 'Delete notification class',
			message: `Delete ${item.event_category}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteNotificationClass(item.id);
			addToast('Notification class deleted', 'success');
			notificationClassesStore.reload();
		} catch (err) {
			addToast(
				err instanceof Error ? err.message : 'Failed to delete notification class',
				'error'
			);
		}
	}

	onMount(() => {
		notificationClassesStore.load();
	});
</script>

<svelte:head>
	<title>Notification Classes | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Notification Classes</h1>
			<p class="text-sm text-muted-foreground">Manage notification classes and event categories.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Notification Class
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
			{ key: 'event_category', label: 'Event Category' },
			{ key: 'nc', label: 'NC' },
			{ key: 'object_description', label: 'Description' },
			{ key: 'meaning', label: 'Meaning' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search notification classes..."
		emptyMessage="No notification classes found. Create your first notification class to get started."
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
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/notification-classes/${item.id}`)}>
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
