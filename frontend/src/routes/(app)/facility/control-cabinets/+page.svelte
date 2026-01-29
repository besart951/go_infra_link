<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { controlCabinetsStore } from '$lib/stores/list/entityStores.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import ControlCabinetForm from '$lib/components/facility/ControlCabinetForm.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { addToast } from '$lib/components/toast.svelte';
	import {
		deleteControlCabinet,
		getControlCabinetDeleteImpact
	} from '$lib/infrastructure/api/facility.adapter.js';

	let showForm = $state(false);
	let editingItem: ControlCabinet | undefined = $state(undefined);

	function handleEdit(item: ControlCabinet) {
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
		controlCabinetsStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	async function handleDelete(item: ControlCabinet) {
		try {
			const impact = await getControlCabinetDeleteImpact(item.id);

			if (impact.sps_controllers_count > 0) {
				const ok1 = await confirm({
					title: 'Delete control cabinet',
					message: `This will also delete ${impact.sps_controllers_count} SPS controller(s). Continue?`,
					confirmText: 'Continue',
					cancelText: 'Cancel',
					variant: 'destructive'
				});
				if (!ok1) return;

				const ok2 = await confirm({
					title: 'Confirm cascading delete',
					message: `This will also delete ${impact.sps_controller_system_types_count} system type link(s), ${impact.field_devices_count} field device(s), and ${impact.bacnet_objects_count} bacnet object(s).`,
					confirmText: 'Delete everything',
					cancelText: 'Cancel',
					variant: 'destructive'
				});
				if (!ok2) return;
			}

			await deleteControlCabinet(item.id);
			addToast('Control cabinet deleted', 'success');
			controlCabinetsStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete control cabinet', 'error');
		}
	}

	onMount(() => {
		controlCabinetsStore.load();
	});
</script>

<svelte:head>
	<title>Control Cabinets | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Control Cabinets</h1>
			<p class="text-sm text-muted-foreground">Manage control cabinets and their configurations.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Control Cabinet
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ControlCabinetForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$controlCabinetsStore}
		columns={[
			{ key: 'cabinet_nr', label: 'Cabinet Nr' },
			{ key: 'building', label: 'Building ID' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search control cabinets..."
		emptyMessage="No control cabinets found. Create your first control cabinet to get started."
		onSearch={(text) => controlCabinetsStore.search(text)}
		onPageChange={(page) => controlCabinetsStore.goToPage(page)}
		onReload={() => controlCabinetsStore.reload()}
	>
		{#snippet rowSnippet(cabinet: ControlCabinet)}
			<Table.Cell class="font-medium">
				<a href="/facility/control-cabinets/{cabinet.id}" class="hover:underline">
					{cabinet.control_cabinet_nr ?? 'N/A'}
				</a>
			</Table.Cell>
			<Table.Cell>{cabinet.building_id}</Table.Cell>
			<Table.Cell>
				{new Date(cabinet.created_at).toLocaleDateString()}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(cabinet)}>
						<Pencil class="size-4" />
					</Button>
					<Button variant="ghost" size="sm" href="/facility/control-cabinets/{cabinet.id}">
						View
					</Button>
					<Button variant="ghost" size="sm" onclick={() => handleDelete(cabinet)}>Delete</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
