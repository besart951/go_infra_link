<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import { ArrowDown, ArrowUp, ChevronDown, ChevronRight, Settings2 } from '@lucide/svelte';
  import { keyboardTableNavigation } from '$lib/actions/keyboardTableNavigation.js';
  import BacnetObjectsEditor from '../bacnet/BacnetObjectsEditor.svelte';
  import FieldDeviceTableRow from './FieldDeviceTableRow.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';
  import type { FieldDevice } from '$lib/domain/facility/index.js';
  import type { TableGroupNode } from '$lib/state/table/TableViewState.svelte.js';
  import type { FieldDeviceGroupKey } from './state/FieldDeviceTableView.svelte.js';

  const t = createTranslator();
  const state = useFieldDeviceState();

  const baseColumnCount = 11;
  const specColumnCount = 11;
  const columnCount = $derived.by(() =>
    state.showSpecifications ? baseColumnCount + specColumnCount : baseColumnCount
  );
  const tableClass = $derived([state.view.tableClass, 'w-max min-w-full table-auto'].join(' '));
  const sortableButtonClass =
    'h-auto w-full min-w-0 cursor-pointer justify-start overflow-hidden p-0 text-left underline-offset-4 hover:underline';
  const sortableLabelClass = 'min-w-0 truncate';

  function getGroupLabelKey(key: FieldDeviceGroupKey): string {
    return (
      state.view.grouping.definitions.find((definition) => definition.key === key)?.labelKey ?? ''
    );
  }
</script>

<div
  use:keyboardTableNavigation
  class="max-w-full min-w-0 overflow-hidden rounded-xl border bg-card shadow-sm"
>
  <Table.Root class={tableClass}>
    <Table.Header>
      <Table.Row>
        <Table.Head resizable={false} class="w-8 max-w-8 min-w-8 px-0! py-1!">
          <div class="flex justify-center">
            <Checkbox
              checked={state.allSelected}
              indeterminate={state.someSelected}
              onCheckedChange={() => state.toggleSelectAll()}
              aria-label={$t('field_device.table.select_all')}
            />
          </div>
        </Table.Head>
        <Table.Head resizable={false} class="w-6 max-w-6 min-w-6 px-0!"></Table.Head>
        <Table.Head minResizeWidth={96} class="w-fit max-w-96 min-w-max text-xs">
          <Button
            type="button"
            variant="ghost"
            class={`${sortableButtonClass} text-xs`}
            onclick={() => void state.toggleSort('sps_system_type')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.sps_system_type')}</span>
            {#if state.sortState('sps_system_type') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('sps_system_type') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-36 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('bmk')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.bmk')}</span>
            {#if state.sortState('bmk') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('bmk') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('description')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.description')}</span>
            {#if state.sortState('description') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('description') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('text_fix')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.text_fix')}</span>
            {#if state.sortState('text_fix') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('text_fix') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={56} class="w-14 max-w-14 min-w-14">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('apparat_nr')}
          >
            <span class={sortableLabelClass}>Nr</span>
            {#if state.sortState('apparat_nr') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('apparat_nr') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-40 min-w-18">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('apparat')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.apparat')}</span>
            {#if state.sortState('apparat') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('apparat') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-40 min-w-18">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void state.toggleSort('system_part')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.system_part')}</span>
            {#if state.sortState('system_part') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('system_part') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head resizable={false} class="w-10 max-w-10 min-w-10">
          <Button
            variant={state.showSpecifications ? 'secondary' : 'ghost'}
            size="sm"
            class="h-7 w-7 p-0"
            onclick={() => void state.toggleSpecifications()}
            title={state.showSpecifications
              ? $t('field_device.table.hide_specifications')
              : $t('field_device.table.show_specifications')}
          >
            <Settings2 class="h-4 w-4" />
          </Button>
        </Table.Head>
        {#if state.showSpecifications}
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_supplier')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.supplier')}</span>
              {#if state.sortState('spec_supplier') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_supplier') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_brand')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.brand')}</span>
              {#if state.sortState('spec_brand') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_brand') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_type')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.type')}</span>
              {#if state.sortState('spec_type') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_type') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_motor_valve')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.motor_valve')}</span>
              {#if state.sortState('spec_motor_valve') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_motor_valve') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_size')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.size')}</span>
              {#if state.sortState('spec_size') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_size') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_install_loc')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.install_location')}</span>
              {#if state.sortState('spec_install_loc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_install_loc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_ph')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.ph')}</span>
              {#if state.sortState('spec_ph') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_ph') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_acdc')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.acdc')}</span>
              {#if state.sortState('spec_acdc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_acdc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_amperage')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.amperage')}</span>
              {#if state.sortState('spec_amperage') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_amperage') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_power')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.power')}</span>
              {#if state.sortState('spec_power') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_power') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void state.toggleSort('spec_rotation')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.rotation')}</span>
              {#if state.sortState('spec_rotation') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_rotation') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
        {/if}
        <Table.Head resizable={false} class="w-20 max-w-20 min-w-20"></Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#snippet renderDevice(device: FieldDevice)}
        <FieldDeviceTableRow {device} />

        {#if state.isBacnetExpanded(device.id)}
          <Table.Row class="bg-muted/30 hover:bg-muted/40">
            <Table.Cell colspan={columnCount} class="p-0">
              {#if state.isBacnetLoading(device.id)}
                <div class="p-4">
                  <Skeleton class="h-10 w-full" />
                </div>
              {:else}
                <BacnetObjectsEditor
                  bacnetObjects={device.bacnet_objects ?? []}
                  pendingEdits={state.editing.getBacnetPendingEdits(device.id) ?? new Map()}
                  fieldErrors={state.editing.getBacnetFieldErrors(device.id) ?? new Map()}
                  clientErrors={state.editing.getBacnetClientErrors(device.id) ?? new Map()}
                  sharedEditors={state.getEditorsForDevice(device.id)}
                  disabled={!state.canUpdateFieldDeviceBacnetObjects()}
                  onEdit={(objectId, field, value) => {
                    state.editing.queueBacnetEdit(device.id, objectId, field, value);
                  }}
                  onUndoField={(objectId, field) => {
                    state.editing.discardBacnetObjectFieldEdit(device.id, objectId, field);
                  }}
                  onUndoRow={(objectId) => {
                    state.editing.discardBacnetObjectEdits(device.id, objectId);
                  }}
                />
              {/if}
            </Table.Cell>
          </Table.Row>
        {/if}
      {/snippet}

      {#snippet renderGroup(group: TableGroupNode<FieldDevice, FieldDeviceGroupKey>)}
        <Table.Row class="bg-muted/35 hover:bg-muted/50">
          <Table.Cell colspan={columnCount} class="p-0">
            <Button
              type="button"
              variant="ghost"
              class="flex h-auto w-full justify-start gap-2 px-3 py-2 text-left font-medium"
              style={`padding-left: ${0.75 + group.level * 1.25}rem`}
              aria-expanded={state.view.grouping.isGroupExpanded(group.id)}
              onclick={() => state.view.grouping.toggleGroupExpansion(group.id)}
            >
              {#if state.view.grouping.isGroupExpanded(group.id)}
                <ChevronDown class="h-4 w-4 text-muted-foreground" />
              {:else}
                <ChevronRight class="h-4 w-4 text-muted-foreground" />
              {/if}
              <span class="text-xs font-semibold text-muted-foreground uppercase">
                {$t(getGroupLabelKey(group.key))}
              </span>
              <span class="truncate">{group.label}</span>
              <span
                class="ml-auto rounded-md bg-background px-1.5 py-0.5 text-xs font-medium text-muted-foreground"
              >
                {$t('field_device.view.group_count', { count: group.count })}
              </span>
            </Button>
          </Table.Cell>
        </Table.Row>

        {#if state.view.grouping.isGroupExpanded(group.id)}
          {#if group.children.length > 0}
            {#each group.children as child (child.id)}
              {@render renderGroup(child)}
            {/each}
          {:else}
            {#each group.items as device (device.id)}
              {@render renderDevice(device)}
            {/each}
          {/if}
        {/if}
      {/snippet}

      {#if state.loading}
        <Table.LoadingRows loading {columnCount} rowCount={8} delayMs={0} />
      {:else if state.items.length === 0}
        <Table.Row>
          <Table.Cell colspan={columnCount} class="h-24 text-center">
            <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
              <p class="font-medium">{$t('field_device.empty.title')}</p>
              {#if state.searchText}
                <p class="text-sm">{$t('field_device.empty.search_hint')}</p>
              {/if}
            </div>
          </Table.Cell>
        </Table.Row>
      {:else if state.view.grouping.isGrouped}
        {#each state.tableGroups as group (group.id)}
          {@render renderGroup(group)}
        {/each}
      {:else}
        {#each state.items as device (device.id)}
          {@render renderDevice(device)}
        {/each}
      {/if}
    </Table.Body>
  </Table.Root>
</div>
