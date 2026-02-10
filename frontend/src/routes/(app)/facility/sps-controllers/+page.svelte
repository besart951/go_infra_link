<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { spsControllersStore } from '$lib/stores/list/entityStores.js';
	import type { SPSController } from '$lib/domain/facility/index.js';
	import SPSControllerForm from '$lib/components/facility/SPSControllerForm.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { deleteSPSController } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: SPSController | undefined = $state(undefined);

	function handleEdit(item: SPSController) {
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
		spsControllersStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	async function handleDelete(item: SPSController) {
		const ok = await confirm({
			title: 'Delete SPS controller',
			message: `Delete ${item.device_name}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteSPSController(item.id);
			addToast('SPS controller deleted', 'success');
			spsControllersStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete SPS controller', 'error');
		}
	}

	async function handleCopy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}

	onMount(() => {
		spsControllersStore.load();
	});
</script>

<svelte:head>
	<title>SPS Controllers | Infra Link</title>
</svelte:head>
<ConfirmDialog />
<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">SPS Controllers</h1>
			<p class="text-sm text-muted-foreground">
				Manage SPS controller devices and their configurations.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New SPS Controller
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SPSControllerForm
			initialData={editingItem}
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$spsControllersStore}
		columns={[
			{ key: 'device_name', label: 'Device Name' },
			{ key: 'ga_device', label: 'GA Device' },
			{ key: 'ip_address', label: 'IP Address' },
			{ key: 'cabinet', label: 'Cabinet' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search SPS controllers..."
		emptyMessage="No SPS controllers found. Create your first SPS controller to get started."
		onSearch={(text) => spsControllersStore.search(text)}
		onPageChange={(page) => spsControllersStore.goToPage(page)}
		onReload={() => spsControllersStore.reload()}
	>
		{#snippet rowSnippet(controller: SPSController)}
			<Table.Cell class="font-medium">
				<a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
					{controller.device_name}
				</a>
			</Table.Cell>
			<Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
			<Table.Cell>
				{#if controller.ip_address}
					<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
						{controller.ip_address}
					</code>
				{:else}
					-
				{/if}
			</Table.Cell>
			<Table.Cell>{controller.control_cabinet_id}</Table.Cell>
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
						<DropdownMenu.Item
							onclick={() => handleCopy(controller.device_name ?? controller.id)}
						>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item
							onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}
						>
							View
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(controller)}>
							Edit
						</DropdownMenu.Item>
						<DropdownMenu.Separator />
						<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(controller)}>
							Delete
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
