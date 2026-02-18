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
	import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	const manageSPSController = new ManageSPSControllerUseCase(spsControllerRepository);
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let showForm = $state(false);
	let editingItem: SPSController | undefined = $state(undefined);
	let cabinetMap = $state(new Map<string, string>());
	const cabinetRequests = new Set<string>();

	function updateCabinetMap(cabinets: ControlCabinet[]) {
		const next = new Map(cabinetMap);
		for (const cabinet of cabinets) {
			next.set(cabinet.id, cabinet.control_cabinet_nr ?? cabinet.id);
		}
		cabinetMap = next;
	}

	function getCabinetLabel(cabinetId: string): string {
		return cabinetMap.get(cabinetId) ?? cabinetId;
	}

	async function ensureCabinetLabels(items: SPSController[]) {
		const uniqueIds = new Set(
			items.map((item) => item.control_cabinet_id).filter((id): id is string => Boolean(id))
		);
		const missingIds = Array.from(uniqueIds).filter(
			(id) => !cabinetMap.has(id) && !cabinetRequests.has(id)
		);

		if (missingIds.length === 0) return;

		missingIds.forEach((id) => cabinetRequests.add(id));

		try {
			const cabinets = await controlCabinetRepository.getBulk(missingIds);
			updateCabinetMap(cabinets);
		} catch (err) {
			console.error('Failed to load control cabinets:', err);
		} finally {
			missingIds.forEach((id) => cabinetRequests.delete(id));
		}
	}

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
			title: $t('common.delete'),
			message: $t('facility.delete_sps_controller_confirm').replace('{name}', item.device_name),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await manageSPSController.delete(item.id);
			addToast($t('facility.sps_controller_deleted'), 'success');
			spsControllersStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('facility.delete_sps_controller_failed'), 'error');
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

	$effect(() => {
		const items = $spsControllersStore.items;
		if (items.length > 0) {
			void ensureCabinetLabels(items);
		}
	});
</script>

<svelte:head>
	<title>{$t('facility.sps_controllers_title')} | Infra Link</title>
</svelte:head>
<ConfirmDialog />
<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.sps_controllers_title')}</h1>
			<p class="text-sm text-muted-foreground">
				{$t('facility.sps_controllers_desc')}
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				{$t('facility.new_sps_controller')}
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
			{ key: 'device_name', label: $t('facility.device_name') },
			{ key: 'ga_device', label: $t('facility.ga_device') },
			{ key: 'ip_address', label: $t('facility.ip_address') },
			{ key: 'cabinet', label: 'Cabinet Nr' },
			{ key: 'actions', label: '', width: 'w-[100px]' }
		]}
		searchPlaceholder={$t('facility.search_sps_controllers')}
		emptyMessage={$t('facility.no_sps_controllers_found')}
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
			<Table.Cell>{getCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
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
						<DropdownMenu.Item onclick={() => handleCopy(controller.device_name ?? controller.id)}>
						{$t('facility.copy')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}>
						{$t('facility.view')}
					</DropdownMenu.Item>
					<DropdownMenu.Item onclick={() => handleEdit(controller)}>{$t('common.edit')}</DropdownMenu.Item>
					<DropdownMenu.Separator />
					<DropdownMenu.Item variant="destructive" onclick={() => handleDelete(controller)}>
						{$t('common.delete')}
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
