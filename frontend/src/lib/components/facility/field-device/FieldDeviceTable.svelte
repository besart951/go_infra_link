<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import { ArrowDown, ArrowUp, Settings2 } from '@lucide/svelte';
  import BacnetObjectsEditor from '../bacnet/BacnetObjectsEditor.svelte';
  import FieldDeviceTableRow from './FieldDeviceTableRow.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useFieldDeviceState();

  const baseColumnCount = 11;
  const specColumnCount = 11;
  const columnCount = $derived.by(() =>
    state.showSpecifications ? baseColumnCount + specColumnCount : baseColumnCount
  );
</script>

<div class="rounded-lg border bg-background">
  <Table.Root class="[&_td]:p-2 [&_th]:h-10 [&_th]:px-2">
    <Table.Header>
      <Table.Row>
        <Table.Head class="w-10">
          <Checkbox
            checked={state.allSelected}
            indeterminate={state.someSelected}
            onCheckedChange={() => state.toggleSelectAll()}
            aria-label={$t('field_device.table.select_all')}
          />
        </Table.Head>
        <Table.Head class="w-10"></Table.Head>
        <Table.Head>
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 text-left underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('sps_system_type')}
          >
            <span>{$t('field_device.table.sps_system_type')}</span>
            {#if state.sortState('sps_system_type') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('sps_system_type') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head>
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('bmk')}
          >
            <span>{$t('field_device.table.bmk')}</span>
            {#if state.sortState('bmk') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('bmk') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head>
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('description')}
          >
            <span>{$t('field_device.table.description')}</span>
            {#if state.sortState('description') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('description') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head>
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('text_fix')}
          >
            <span>{$t('field_device.table.text_fix')}</span>
            {#if state.sortState('text_fix') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('text_fix') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head class="w-24">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('apparat_nr')}
          >
            <span>{$t('field_device.table.apparat_nr')}</span>
            {#if state.sortState('apparat_nr') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('apparat_nr') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head class="w-48">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('apparat')}
          >
            <span>{$t('field_device.table.apparat')}</span>
            {#if state.sortState('apparat') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('apparat') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head class="w-48">
          <button
            type="button"
            class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
            onclick={() => void state.toggleSort('system_part')}
          >
            <span>{$t('field_device.table.system_part')}</span>
            {#if state.sortState('system_part') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if state.sortState('system_part') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </button>
        </Table.Head>
        <Table.Head class="w-10">
          <Button
            variant={state.showSpecifications ? 'secondary' : 'ghost'}
            size="sm"
            class="h-7 w-7 p-0"
            onclick={() => state.toggleSpecifications()}
            title={state.showSpecifications
              ? $t('field_device.table.hide_specifications')
              : $t('field_device.table.show_specifications')}
          >
            <Settings2 class="h-4 w-4" />
          </Button>
        </Table.Head>
        {#if state.showSpecifications}
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_supplier')}
            >
              <span>{$t('field_device.table.supplier')}</span>
              {#if state.sortState('spec_supplier') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_supplier') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_brand')}
            >
              <span>{$t('field_device.table.brand')}</span>
              {#if state.sortState('spec_brand') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_brand') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_type')}
            >
              <span>{$t('field_device.table.type')}</span>
              {#if state.sortState('spec_type') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_type') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_motor_valve')}
            >
              <span>{$t('field_device.table.motor_valve')}</span>
              {#if state.sortState('spec_motor_valve') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_motor_valve') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_size')}
            >
              <span>{$t('field_device.table.size')}</span>
              {#if state.sortState('spec_size') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_size') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_install_loc')}
            >
              <span>{$t('field_device.table.install_location')}</span>
              {#if state.sortState('spec_install_loc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_install_loc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_ph')}
            >
              <span>{$t('field_device.table.ph')}</span>
              {#if state.sortState('spec_ph') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_ph') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_acdc')}
            >
              <span>{$t('field_device.table.acdc')}</span>
              {#if state.sortState('spec_acdc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_acdc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_amperage')}
            >
              <span>{$t('field_device.table.amperage')}</span>
              {#if state.sortState('spec_amperage') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_amperage') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_power')}
            >
              <span>{$t('field_device.table.power')}</span>
              {#if state.sortState('spec_power') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_power') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
          <Table.Head class="text-xs">
            <button
              type="button"
              class="inline-flex cursor-pointer items-center gap-1 underline-offset-4 hover:underline"
              onclick={() => void state.toggleSort('spec_rotation')}
            >
              <span>{$t('field_device.table.rotation')}</span>
              {#if state.sortState('spec_rotation') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if state.sortState('spec_rotation') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </button>
          </Table.Head>
        {/if}
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#if state.loading && state.items.length === 0}
        {#each Array(5) as _}
          <Table.Row>
            {#each Array(columnCount) as _}
              <Table.Cell>
                <Skeleton class="h-8 w-full" />
              </Table.Cell>
            {/each}
          </Table.Row>
        {/each}
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
      {:else}
        {#each state.items as device (device.id)}
          <FieldDeviceTableRow {device} />

          {#if state.isBacnetExpanded(device.id)}
            <Table.Row
              class="bg-purple-50/50 hover:bg-purple-50/70 dark:bg-purple-950/20 dark:hover:bg-purple-950/30"
            >
              <Table.Cell colspan={columnCount} class="p-0">
                <BacnetObjectsEditor
                  bacnetObjects={device.bacnet_objects ?? []}
                  pendingEdits={state.editing.getBacnetPendingEdits(device.id) ?? new Map()}
                  fieldErrors={state.editing.getBacnetFieldErrors(device.id) ?? new Map()}
                  clientErrors={state.editing.getBacnetClientErrors(device.id) ?? new Map()}
                  disabled={!canPerform('update', 'fielddevice')}
                  onEdit={(objectId, field, value) => {
                    state.editing.queueBacnetEdit(device.id, objectId, field, value);
                  }}
                />
              </Table.Cell>
            </Table.Row>
          {/if}
        {/each}
      {/if}
    </Table.Body>
  </Table.Root>
</div>
