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
	import { buildingsStore } from '$lib/stores/list/entityStores.js';
	import type { Building } from '$lib/domain/facility/index.js';
	import BuildingForm from '$lib/components/facility/BuildingForm.svelte';
	import { deleteBuilding } from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingBuilding: Building | undefined = $state(undefined);

	function handleEdit(building: Building) {
		editingBuilding = building;
		showForm = true;
	}

	function handleCreate() {
		editingBuilding = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingBuilding = undefined;
		buildingsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingBuilding = undefined;
	}

	async function handleCopy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}

	async function handleDelete(building: Building) {
		const ok = await confirm({
			title: 'Delete building',
			message: `Delete ${building.iws_code}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await deleteBuilding(building.id);
			addToast('Building deleted', 'success');
			buildingsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete building', 'error');
		}
	}

	onMount(() => {
		buildingsStore.load();
	});
</script>

<svelte:head>
	<title>Buildings | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Buildings</h1>
			<p class="text-sm text-muted-foreground">Manage building infrastructure and IWS codes.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Building
			</Button>
		{/if}
	</div>

	{#if showForm}
		<BuildingForm initialData={editingBuilding} onSuccess={handleSuccess} onCancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$buildingsStore}
		columns={[
			{ key: 'iws_code', label: 'IWS Code' },
			{ key: 'building_group', label: 'Building Group' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search buildings..."
		emptyMessage="No buildings found. Create your first building to get started."
		onSearch={(text) => buildingsStore.search(text)}
		onPageChange={(page) => buildingsStore.goToPage(page)}
		onReload={() => buildingsStore.reload()}
	>
		{#snippet rowSnippet(building: Building)}
			<Table.Cell class="font-medium">
				<a href="/facility/buildings/{building.id}" class="hover:underline">
					{building.iws_code}
				</a>
			</Table.Cell>
			<Table.Cell>{building.building_group}</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(building.iws_code)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/buildings/${building.id}`)}>
							View
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(building)}>
							Edit
						</DropdownMenu.Item>
						<DropdownMenu.Separator />
						<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(building)}>
							Delete
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
