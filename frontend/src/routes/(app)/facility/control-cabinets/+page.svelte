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
	import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	const manageControlCabinet = new ManageControlCabinetUseCase(controlCabinetRepository);
	import type { Building } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let showForm = $state(false);
	let editingItem: ControlCabinet | undefined = $state(undefined);
	let buildingMap = $state(new Map<string, string>());
	const buildingRequests = new Set<string>();

	function formatBuildingLabel(building: Building): string {
		return `${building.iws_code}-${building.building_group}`;
	}

	function getBuildingLabel(buildingId: string): string {
		return buildingMap.get(buildingId) ?? buildingId;
	}

	function updateBuildingMap(buildings: Building[]) {
		const next = new Map(buildingMap);
		for (const building of buildings) {
			next.set(building.id, formatBuildingLabel(building));
		}
		buildingMap = next;
	}

	async function ensureBuildingLabels(items: ControlCabinet[]) {
		const uniqueIds = new Set(
			items.map((item) => item.building_id).filter((id): id is string => Boolean(id))
		);
		const missingIds = Array.from(uniqueIds).filter(
			(id) => !buildingMap.has(id) && !buildingRequests.has(id)
		);

		if (missingIds.length === 0) return;

		missingIds.forEach((id) => buildingRequests.add(id));

		try {
			const buildings = await buildingRepository.getBulk(missingIds);
			updateBuildingMap(buildings);
		} catch (err) {
			console.error('Failed to load buildings:', err);
		} finally {
			missingIds.forEach((id) => buildingRequests.delete(id));
		}
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
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}

	async function handleDelete(item: ControlCabinet) {
		try {
			const impact = await manageControlCabinet.getDeleteImpact(item.id);

			if (impact.sps_controllers_count > 0) {
				const ok1 = await confirm({
					title: $t('facility.delete_control_cabinet_confirm'),
					message: $t('facility.delete_control_cabinet_message').replace(
						'{count}',
						impact.sps_controllers_count.toString()
					),
					confirmText: $t('common.confirm'),
					cancelText: $t('common.cancel'),
					variant: 'destructive'
				});
				if (!ok1) return;

				const ok2 = await confirm({
					title: $t('facility.confirm_cascading_delete'),
					message: $t('facility.cascading_delete_message')
						.replace('{systemTypes}', impact.sps_controller_system_types_count.toString())
						.replace('{fieldDevices}', impact.field_devices_count.toString())
						.replace('{bacnetObjects}', impact.bacnet_objects_count.toString()),
					confirmText: $t('facility.delete_everything'),
					cancelText: $t('common.cancel'),
					variant: 'destructive'
				});
				if (!ok2) return;
			}

			await manageControlCabinet.delete(item.id);
			addToast($t('facility.control_cabinet_deleted'), 'success');
			controlCabinetsStore.reload();
		} catch (err) {
			addToast(
				err instanceof Error ? err.message : $t('facility.delete_control_cabinet_failed'),
				'error'
			);
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
	});

	$effect(() => {
		const items = $controlCabinetsStore.items;
		if (items.length > 0) {
			void ensureBuildingLabels(items);
		}
	});
</script>

<svelte:head>
	<title>{$t('facility.control_cabinets_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.control_cabinets_title')}</h1>
			<p class="text-sm text-muted-foreground">{$t('facility.control_cabinets_desc')}</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_control_cabinet')}
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
			{ key: 'building', label: $t('facility.building') },
			{ key: 'cabinet_nr', label: 'Cabinet Nr' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_control_cabinets')}
		emptyMessage={$t('facility.no_control_cabinets_found')}
		onSearch={(text) => controlCabinetsStore.search(text)}
		onPageChange={(page) => controlCabinetsStore.goToPage(page)}
		onReload={() => controlCabinetsStore.reload()}
	>
		{#snippet rowSnippet(cabinet: ControlCabinet)}
			<Table.Cell class="font-medium">
				<a href="/facility/control-cabinets/{cabinet.id}" class="hover:underline">
					{cabinet.control_cabinet_nr ?? $t('common.not_available')}
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
							{$t('facility.copy')}
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => goto(`/facility/control-cabinets/${cabinet.id}`)}>
							{$t('facility.view')}
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => handleEdit(cabinet)}
							>{$t('common.edit')}</DropdownMenu.Item
						>
						<DropdownMenu.Separator />
						<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(cabinet)}>
							{$t('common.delete')}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
