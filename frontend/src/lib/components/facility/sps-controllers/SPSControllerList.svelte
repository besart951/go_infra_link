<script lang="ts">
  import { goto } from '$app/navigation';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import type { SPSController } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { Plus } from '@lucide/svelte';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { useSPSControllerState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useSPSControllerState();

  const columns = $derived.by(() =>
    state.isProjectContext
      ? [
          { key: 'device_name', label: $t('projects.sps_controllers.columns.device_name') },
          { key: 'ga_device', label: $t('projects.sps_controllers.columns.ga_device') },
          { key: 'ip_address', label: $t('projects.sps_controllers.columns.ip_address') },
          { key: 'cabinet', label: $t('projects.sps_controllers.columns.cabinet') },
          { key: 'created', label: $t('projects.sps_controllers.columns.created') },
          { key: 'actions', label: '', width: 'w-[100px]' }
        ]
      : [
          { key: 'device_name', label: $t('facility.device_name') },
          { key: 'cabinet', label: 'Cabinet Nr' },
          { key: 'ga_device', label: $t('facility.ga_device') },
          { key: 'ip_address', label: $t('facility.ip_address') },
          { key: 'system_types', label: $t('facility.system_types') },
          { key: 'actions', label: '', width: 'w-[100px]' }
        ]
  );

  const searchPlaceholder = $derived.by(() =>
    state.isProjectContext
      ? $t('projects.sps_controllers.search_placeholder')
      : $t('facility.search_sps_controllers')
  );

  const emptyMessage = $derived.by(() =>
    state.isProjectContext
      ? $t('projects.sps_controllers.empty')
      : $t('facility.no_sps_controllers_found')
  );

  const newLabel = $derived.by(() =>
    state.isProjectContext ? $t('projects.sps_controllers.new') : $t('facility.new_sps_controller')
  );
</script>

<div class="flex flex-col gap-4">
  <div class="flex flex-wrap items-center justify-end gap-2">
    {#if !state.showForm && state.canCreateSPSController()}
      <Button onclick={() => state.openCreateForm()}>
        <Plus class="mr-2 size-4" />
        {newLabel}
      </Button>
    {/if}
  </div>

  {#if state.showForm}
    <SPSControllerForm
      initialData={state.editingItem}
      projectId={state.projectId}
      controlCabinetRefreshKey={state.controlCabinetRefreshKey}
      onSuccess={(controller) => void state.handleFormSuccess(controller)}
      onCancel={() => state.cancelForm()}
    />
  {/if}

  <Tooltip.Provider>
    <PaginatedList
      {state}
      {columns}
      {searchPlaceholder}
      {emptyMessage}
      onSearch={(text) => void state.search(text)}
      onPageChange={(page) => void state.goToPage(page)}
      onReload={() => void state.reload()}
    >
      {#snippet rowSnippet(controller: SPSController)}
        {@const systemTypes = state.getSystemTypes(controller.id)}
        <Table.Cell class="font-medium">
          <a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
            {controller.device_name}
          </a>
        </Table.Cell>

        {#if state.isProjectContext}
          <Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
          <Table.Cell>
            {#if controller.ip_address}
              <code class="rounded-md bg-muted px-1.5 py-0.5 text-sm">
                {controller.ip_address}
              </code>
            {:else}
              -
            {/if}
          </Table.Cell>
          <Table.Cell>{state.getCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
          <Table.Cell>{new Date(controller.created_at).toLocaleDateString()}</Table.Cell>
        {:else}
          <Table.Cell>{state.getCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
          <Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
          <Table.Cell>
            {#if controller.ip_address}
              <code class="rounded-md bg-muted px-1.5 py-0.5 text-sm">
                {controller.ip_address}
              </code>
            {:else}
              -
            {/if}
          </Table.Cell>
          <Table.Cell>
            {#if !state.hasLoadedSystemTypes(controller.id)}
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
                        <div
                          class="rounded-md border border-border/60 bg-muted/20 px-3 py-2 text-sm"
                        >
                          <p class="font-medium text-foreground">
                            {state.formatSystemTypeTitle(systemType)}
                          </p>
                          {#if state.formatSystemTypeMeta(systemType)}
                            <p class="text-xs text-muted-foreground">
                              {state.formatSystemTypeMeta(systemType)}
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
        {/if}

        <Table.Cell class="text-right">
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              {#snippet child({ props })}
                <Button variant="ghost" size="icon" {...props}>
                  <EllipsisIcon class="size-4" />
                </Button>
              {/snippet}
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end" class="w-44">
              {#if !state.isProjectContext}
                <DropdownMenu.Item
                  onclick={() =>
                    void state.copyToClipboard(controller.device_name ?? controller.id)}
                >
                  {$t('facility.copy')}
                </DropdownMenu.Item>
              {/if}
              {#if state.canCreateSPSController()}
                <DropdownMenu.Item onclick={() => void state.duplicateSPSController(controller)}>
                  {$t('facility.duplicate')}
                </DropdownMenu.Item>
              {/if}
              {#if state.canReadSPSController()}
                <DropdownMenu.Item
                  onclick={() => goto(`/facility/sps-controllers/${controller.id}`)}
                >
                  {$t('common.view')}
                </DropdownMenu.Item>
              {/if}
              {#if state.canUpdateSPSController()}
                <DropdownMenu.Item onclick={() => state.editSPSController(controller)}>
                  {$t('common.edit')}
                </DropdownMenu.Item>
              {/if}
              {#if state.canDeleteSPSController()}
                <DropdownMenu.Separator />
                <DropdownMenu.Item
                  variant="destructive"
                  onclick={() => void state.deleteSPSController(controller)}
                >
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
