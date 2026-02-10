<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { controlCabinetsStore } from '$lib/stores/list/entityStores.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import ControlCabinetForm from '$lib/components/facility/ControlCabinetForm.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import {
		deleteControlCabinet,
		getControlCabinetDeleteImpact,
		listBuildings
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { Building } from '$lib/domain/facility/index.js';

	let showForm = $state(false);
	let editingItem: ControlCabinet | undefined = $state(undefined);
	let buildingMap = $state(new Map<string, string>());

	function formatBuildingLabel(building: Building): string {
		return `${building.iws_code}-${building.building_group}`;
	}

	async function loadBuildingMap() {
		try {
			const res = await listBuildings({ page: 1, limit: 1000 });
			const next = new Map<string, string>();
			for (const building of res.items || []) {
				next.set(building.id, formatBuildingLabel(building));
			}
			buildingMap = next;
		} catch (err) {
			console.error('Failed to load buildings:', err);
		}
	}

	function getBuildingLabel(buildingId: string): string {
		return buildingMap.get(buildingId) ?? buildingId;
	}

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
		loadBuildingMap();
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

	async function handleCopy(value: string) {
		try {
			await navigator.clipboard.writeText(value);
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
		}
	}

	onMount(() => {
		controlCabinetsStore.load();
		loadBuildingMap();
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
			onSuccess={handleSuccess}
			onCancel={handleCancel}
		/>
	{/if}

	<PaginatedList
		state={$controlCabinetsStore}
		columns={[
			{ key: 'cabinet_nr', label: 'Cabinet Nr' },
			{ key: 'building', label: 'Building' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
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
			<Table.Cell>{getBuildingLabel(cabinet.building_id)}</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(cabinet.control_cabinet_nr ?? cabinet.id)}>
							Copy
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/control-cabinets/${cabinet.id}`)}>
							View
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(cabinet)}>Edit</DropdownMenu.Item>
						<DropdownMenu.Separator />
						<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(cabinet)}>
							Delete
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
