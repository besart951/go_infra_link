<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { Plus } from '@lucide/svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import { spsControllersStore } from '$lib/stores/list/entityStores.js';
  import type {
    ControlCabinet,
    SPSController,
    SPSControllerSystemType
  } from '$lib/domain/facility/index.js';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { canPerform } from '$lib/utils/permissions.js';
  const manageSPSController = new ManageSPSControllerUseCase(spsControllerRepository);
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  let showForm = $state(false);
  let editingItem: SPSController | undefined = $state(undefined);
  let cabinetMap = $state(new Map<string, string>());
  let systemTypesByController = $state<Record<string, SPSControllerSystemType[]>>({});
  const cabinetRequests = new Set<string>();
  const systemTypeRequests = new Set<string>();

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

  function hasLoadedSystemTypes(controllerId: string): boolean {
    return Object.prototype.hasOwnProperty.call(systemTypesByController, controllerId);
  }

  function getSystemTypes(controllerId: string): SPSControllerSystemType[] {
    return systemTypesByController[controllerId] ?? [];
  }

  function formatSystemTypeTitle(systemType: SPSControllerSystemType): string {
    return systemType.system_type_name ?? systemType.system_type_id;
  }

  function formatSystemTypeMeta(systemType: SPSControllerSystemType): string {
    const parts: string[] = [];
    if (systemType.number != null) {
      parts.push(`${$t('facility.control_cabinet_detail.number')}: ${systemType.number}`);
    }
    if (systemType.document_name) {
      parts.push(systemType.document_name);
    }
    return parts.join(' • ');
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

  async function ensureSystemTypes(items: SPSController[]) {
    const missingIds = Array.from(new Set(items.map((item) => item.id))).filter(
      (id) => !hasLoadedSystemTypes(id) && !systemTypeRequests.has(id)
    );

    if (missingIds.length === 0) return;

    missingIds.forEach((id) => systemTypeRequests.add(id));

    try {
      const results = await Promise.allSettled(
        missingIds.map(async (id) => {
          const response = await spsControllerSystemTypeRepository.list({
            pagination: { page: 1, pageSize: 1000 },
            search: { text: '' },
            filters: { sps_controller_id: id }
          });

          return { id, items: response.items };
        })
      );

      const next = { ...systemTypesByController };

      for (const result of results) {
        if (result.status === 'fulfilled') {
          next[result.value.id] = result.value.items;
          continue;
        }

        console.error('Failed to load SPS controller system types:', result.reason);
      }

      systemTypesByController = next;
    } finally {
      missingIds.forEach((id) => systemTypeRequests.delete(id));
    }
  }

  function reloadControllers() {
    systemTypesByController = {};
    spsControllersStore.reload();
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
    reloadControllers();
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
      reloadControllers();
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : $t('facility.delete_sps_controller_failed'),
        'error'
      );
    }
  }

  async function handleDuplicate(item: SPSController) {
    try {
      await manageSPSController.copy(item.id);
      addToast($t('facility.sps_controller_copied'), 'success');
      reloadControllers();
    } catch (err) {
      addToast(err instanceof Error ? err.message : $t('facility.copy_failed'), 'error');
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
      void ensureSystemTypes(items);
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
    {#if !showForm && canPerform('create', 'spscontroller')}
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

  <Tooltip.Provider>
    <PaginatedList
      state={$spsControllersStore}
      columns={[
        { key: 'device_name', label: $t('facility.device_name') },
        { key: 'cabinet', label: 'Cabinet Nr' },
        { key: 'ga_device', label: $t('facility.ga_device') },
        { key: 'ip_address', label: $t('facility.ip_address') },
        { key: 'system_types', label: $t('facility.system_types') },
        { key: 'actions', label: '', width: 'w-[100px]' }
      ]}
      searchPlaceholder={$t('facility.search_sps_controllers')}
      emptyMessage={$t('facility.no_sps_controllers_found')}
      onSearch={(text) => spsControllersStore.search(text)}
      onPageChange={(page) => spsControllersStore.goToPage(page)}
      onReload={reloadControllers}
    >
      {#snippet rowSnippet(controller: SPSController)}
        {@const systemTypes = getSystemTypes(controller.id)}
        <Table.Cell class="font-medium">
          <a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
            {controller.device_name}
          </a>
        </Table.Cell>
        <Table.Cell>{getCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
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
        <Table.Cell>
          {#if !hasLoadedSystemTypes(controller.id)}
            <Badge variant="outline">...</Badge>
          {:else if systemTypes.length === 0}
            <Badge variant="outline">0</Badge>
          {:else}
            <Tooltip.Root>
              <Tooltip.Trigger class="inline-flex">
                <Badge variant="secondary" class="cursor-help">
                  {systemTypes.length}
                </Badge>
              </Tooltip.Trigger>
              <Tooltip.Content class="max-h-80 max-w-sm overflow-y-auto">
                <div class="space-y-3">
                  <div>
                    <p class="font-medium">{controller.device_name}</p>
                    <p class="text-xs text-muted-foreground">
                      {systemTypes.length}
                      {$t('facility.system_types')}
                    </p>
                  </div>
                  <div class="space-y-2">
                    {#each systemTypes as systemType (systemType.id)}
                      <div class="rounded-md border border-border/60 bg-muted/20 px-3 py-2 text-sm">
                        <p class="font-medium text-foreground">
                          {formatSystemTypeTitle(systemType)}
                        </p>
                        {#if formatSystemTypeMeta(systemType)}
                          <p class="text-xs text-muted-foreground">
                            {formatSystemTypeMeta(systemType)}
                          </p>
                        {/if}
                      </div>
                    {/each}
                  </div>
                </div>
              </Tooltip.Content>
            </Tooltip.Root>
          {/if}
        </Table.Cell>
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
                {$t('facility.copy')}
              </DropdownMenu.Item>
              {#if canPerform('create', 'spscontroller')}
                <DropdownMenu.Item onclick={() => handleDuplicate(controller)}>
                  {$t('facility.duplicate')}
                </DropdownMenu.Item>
              {/if}
              <DropdownMenu.Item onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}>
                {$t('facility.view')}
              </DropdownMenu.Item>
              {#if canPerform('update', 'spscontroller')}
                <DropdownMenu.Item onclick={() => handleEdit(controller)}
                  >{$t('common.edit')}</DropdownMenu.Item
                >
              {/if}
              {#if canPerform('delete', 'spscontroller')}
                <DropdownMenu.Separator />
                <DropdownMenu.Item variant="destructive" onclick={() => handleDelete(controller)}>
                  {$t('common.delete')}
                </DropdownMenu.Item>
              {/if}
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </Table.Cell>
      {/snippet}
    </PaginatedList>
  </Tooltip.Provider>
</div>
