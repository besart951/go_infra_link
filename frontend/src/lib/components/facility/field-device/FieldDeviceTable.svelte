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
  const tableState = useFieldDeviceState();

  const ROW_HEIGHT = 52;
  const BACNET_ROW_HEIGHT = 280;
  const OVERSCAN = 4;

  const baseColumnCount = 11;
  const specColumnCount = 11;
  const columnCount = $derived.by(() =>
    tableState.showSpecifications ? baseColumnCount + specColumnCount : baseColumnCount
  );
  const tableClass = $derived([tableState.view.tableClass, 'w-max min-w-full table-auto'].join(' '));
  const sortableButtonClass =
    'h-auto w-full min-w-0 cursor-pointer justify-start overflow-hidden p-0 text-left underline-offset-4 hover:underline';
  const sortableLabelClass = 'min-w-0 truncate';
  let tableBodyRef: HTMLElement | null = $state(null);
  let tableBodyScrollTop = $state(0);
  let tableBodyHeight = $state(640);
  let lastSelectedDeviceId = $state<string | null>(null);

  type FieldDeviceRenderRow =
    | {
        kind: 'group';
        key: string;
        index: number;
        group: TableGroupNode<FieldDevice, FieldDeviceGroupKey>;
      }
    | {
        kind: 'device';
        key: string;
        index: number;
        device: FieldDevice;
      }
    | {
        kind: 'bacnet';
        key: string;
        index: number;
        device: FieldDevice;
      };

  function isInteractiveRowTarget(target: EventTarget | null): boolean {
    if (!(target instanceof Element)) return false;

    return target.closest(
      'a, button, input, select, textarea, option, label, [data-keyboard-table-cell], .editable-cell-display, .editable-cell-editor'
    ) !== null;
  }

  function handleDeviceRowClick(row: Extract<FieldDeviceRenderRow, { kind: 'device' }>, event: MouseEvent): void {
    const clickedId = row.device.id;
    const isAdditive = event.ctrlKey || event.metaKey;
    const isRange = event.shiftKey && lastSelectedDeviceId !== null;

    if (isRange) {
      const startIndex = allRenderRows.findIndex(
        (entry) => entry.kind === 'device' && entry.device.id === lastSelectedDeviceId
      );
      if (startIndex >= 0) {
        const from = Math.min(startIndex, row.index);
        const to = Math.max(startIndex, row.index);
        const next = new Set<string>();

        for (let i = from; i <= to; i++) {
          const candidate = allRenderRows[i];
          if (candidate && candidate.kind === 'device') {
            next.add(candidate.device.id);
          }
        }

        tableState.setSelectedIds(next);
        lastSelectedDeviceId = clickedId;
        return;
      }
    }

    if (isAdditive) {
      const next = new Set(tableState.selectedIds);
      if (next.has(clickedId)) {
        next.delete(clickedId);
      } else {
        next.add(clickedId);
      }
      tableState.setSelectedIds(next);
      lastSelectedDeviceId = clickedId;
      return;
    }

    if (tableState.selectedIds.has(clickedId)) {
      tableState.setSelectedIds(new Set());
      lastSelectedDeviceId = null;
      return;
    }

    tableState.setSelectedIds(new Set([clickedId]));
    lastSelectedDeviceId = clickedId;
  }

  $effect(() => {
    const root = tableBodyRef;
    if (!root) return;

    const handleResize = () => {
      tableBodyHeight = root.clientHeight;
    };
    const handleScroll = () => {
      tableBodyScrollTop = root.scrollTop;
    };

    handleResize();
    handleScroll();

    const resizeObserver = new ResizeObserver(() => {
      handleResize();
    });
    resizeObserver.observe(root);
    root.addEventListener('scroll', handleScroll, { passive: true });

    return () => {
      resizeObserver.disconnect();
      root.removeEventListener('scroll', handleScroll);
    };
  });

  function rowHeightFor(row: FieldDeviceRenderRow): number {
    return row.kind === 'bacnet' ? BACNET_ROW_HEIGHT : ROW_HEIGHT;
  }

  function addDeviceRow(accumulator: FieldDeviceRenderRow[], device: FieldDevice): void {
    accumulator.push({
      kind: 'device',
      index: accumulator.length,
      key: `d:${device.id}`,
      device
    });
    if (tableState.isBacnetExpanded(device.id)) {
      accumulator.push({
        kind: 'bacnet',
        index: accumulator.length,
        key: `b:${device.id}`,
        device
      });
    }
  }

  function addGroupRows(
    accumulator: FieldDeviceRenderRow[],
    group: TableGroupNode<FieldDevice, FieldDeviceGroupKey>
  ): void {
    accumulator.push({ kind: 'group', index: accumulator.length, key: `g:${group.id}`, group });
    if (!tableState.view.grouping.isGroupExpanded(group.id)) {
      return;
    }

    if (group.children.length > 0) {
      for (const child of group.children) {
        addGroupRows(accumulator, child);
      }
      return;
    }

    for (const device of group.items) {
      addDeviceRow(accumulator, device);
    }
  }

  const allRenderRows = $derived.by(() => {
    const rows: FieldDeviceRenderRow[] = [];

    if (tableState.view.grouping.isGrouped) {
      for (const group of tableState.tableGroups) {
        addGroupRows(rows, group);
      }
    } else {
      for (const device of tableState.items) {
        addDeviceRow(rows, device);
      }
    }

    return rows;
  });

  const renderWindow = $derived.by(() => {
    const rows = allRenderRows;
    const itemCount = rows.length;

    if (itemCount === 0 || tableBodyHeight <= 0) {
      return {
        rows: [] as FieldDeviceRenderRow[],
        topSpacerHeight: 0,
        bottomSpacerHeight: 0
      };
    }

    let start = 0;
    let consumedHeight = 0;
    while (start < itemCount && consumedHeight + rowHeightFor(rows[start]) <= tableBodyScrollTop) {
      consumedHeight += rowHeightFor(rows[start]);
      start++;
    }

    let end = start;
    let viewportHeight = consumedHeight + tableBodyHeight;
    while (end < itemCount && consumedHeight < viewportHeight) {
      consumedHeight += rowHeightFor(rows[end]);
      end++;
    }

    const windowStart = Math.max(0, start - OVERSCAN);
    const windowEnd = Math.min(itemCount, end + OVERSCAN);

    let topSpacerHeight = 0;
    let bottomSpacerHeight = 0;
    for (let i = 0; i < windowStart; i++) {
      topSpacerHeight += rowHeightFor(rows[i]);
    }
    for (let i = windowEnd; i < itemCount; i++) {
      bottomSpacerHeight += rowHeightFor(rows[i]);
    }

    return {
      rows: rows.slice(windowStart, windowEnd),
      topSpacerHeight,
      bottomSpacerHeight
    };
  });

  const apparatNrSuggestionByDeviceId = $derived.by(() => {
    const map = new Map<string, number | undefined>();
    for (const row of renderWindow.rows) {
      if (row.kind !== 'device') continue;
      map.set(
        row.device.id,
        tableState.editing.getFieldSuggestion(row.device.id, 'apparat_nr', tableState.items)
      );
    }
    return map;
  });

  function getGroupLabelKey(key: FieldDeviceGroupKey): string {
    return (
      tableState.view.grouping.definitions.find((definition) => definition.key === key)?.labelKey ?? ''
    );
  }
</script>

<div
  use:keyboardTableNavigation
  bind:this={tableBodyRef}
  class="max-w-full min-w-0 overflow-auto rounded-xl border bg-card shadow-sm max-h-[75vh]"
>
  <Table.Root class={tableClass}>
    <Table.Header>
      <Table.Row>
        <Table.Head resizable={false} class="w-8 max-w-8 min-w-8 px-0! py-1!">
          <div class="flex justify-center">
            <Checkbox
              checked={tableState.allSelected}
              indeterminate={tableState.someSelected}
              onCheckedChange={() => tableState.toggleSelectAll()}
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
            onclick={() => void tableState.toggleSort('sps_system_type')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.sps_system_type')}</span>
            {#if tableState.sortState('sps_system_type') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('sps_system_type') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-36 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('bmk')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.bmk')}</span>
            {#if tableState.sortState('bmk') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('bmk') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('description')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.description')}</span>
            {#if tableState.sortState('description') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('description') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('text_fix')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.text_fix')}</span>
            {#if tableState.sortState('text_fix') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('text_fix') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={56} class="w-14 max-w-14 min-w-14">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('apparat_nr')}
          >
            <span class={sortableLabelClass}>Nr</span>
            {#if tableState.sortState('apparat_nr') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('apparat_nr') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-40 min-w-18">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('apparat')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.apparat')}</span>
            {#if tableState.sortState('apparat') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('apparat') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head minResizeWidth={72} class="w-fit max-w-40 min-w-18">
          <Button
            type="button"
            variant="ghost"
            class={sortableButtonClass}
            onclick={() => void tableState.toggleSort('system_part')}
          >
            <span class={sortableLabelClass}>{$t('field_device.table.system_part')}</span>
            {#if tableState.sortState('system_part') === 'asc'}
              <ArrowUp class="h-3 w-3" />
            {:else if tableState.sortState('system_part') === 'desc'}
              <ArrowDown class="h-3 w-3" />
            {/if}
          </Button>
        </Table.Head>
        <Table.Head resizable={false} class="w-10 max-w-10 min-w-10">
          <Button
            variant={tableState.showSpecifications ? 'secondary' : 'ghost'}
            size="sm"
            class="h-7 w-7 p-0"
            onclick={() => void tableState.toggleSpecifications()}
            title={tableState.showSpecifications
              ? $t('field_device.table.hide_specifications')
              : $t('field_device.table.show_specifications')}
          >
            <Settings2 class="h-4 w-4" />
          </Button>
        </Table.Head>
        {#if tableState.showSpecifications}
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_supplier')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.supplier')}</span>
              {#if tableState.sortState('spec_supplier') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_supplier') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_brand')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.brand')}</span>
              {#if tableState.sortState('spec_brand') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_brand') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_type')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.type')}</span>
              {#if tableState.sortState('spec_type') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_type') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-64 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_motor_valve')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.motor_valve')}</span>
              {#if tableState.sortState('spec_motor_valve') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_motor_valve') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_size')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.size')}</span>
              {#if tableState.sortState('spec_size') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_size') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={96} class="w-fit max-w-80 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_install_loc')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.install_location')}</span>
              {#if tableState.sortState('spec_install_loc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_install_loc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_ph')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.ph')}</span>
              {#if tableState.sortState('spec_ph') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_ph') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={64} class="w-fit max-w-24 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_acdc')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.acdc')}</span>
              {#if tableState.sortState('spec_acdc') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_acdc') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_amperage')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.amperage')}</span>
              {#if tableState.sortState('spec_amperage') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_amperage') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_power')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.power')}</span>
              {#if tableState.sortState('spec_power') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_power') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
          <Table.Head minResizeWidth={72} class="w-fit max-w-32 min-w-max text-xs">
            <Button
              type="button"
              variant="ghost"
              class={sortableButtonClass}
              onclick={() => void tableState.toggleSort('spec_rotation')}
            >
              <span class={sortableLabelClass}>{$t('field_device.table.rotation')}</span>
              {#if tableState.sortState('spec_rotation') === 'asc'}
                <ArrowUp class="h-3 w-3" />
              {:else if tableState.sortState('spec_rotation') === 'desc'}
                <ArrowDown class="h-3 w-3" />
              {/if}
            </Button>
          </Table.Head>
        {/if}
        <Table.Head resizable={false} class="w-20 max-w-20 min-w-20"></Table.Head>
      </Table.Row>
    </Table.Header>
    <Table.Body>
      {#snippet renderDevice(row: Extract<FieldDeviceRenderRow, { kind: 'device' }>) }
        <FieldDeviceTableRow
          device={row.device}
          apparatNrSuggestion={apparatNrSuggestionByDeviceId.get(row.device.id)}
          onRowClick={(event) => {
            if (isInteractiveRowTarget(event.target)) return;
            handleDeviceRowClick(row, event);
          }}
        />
      {/snippet}

      {#snippet renderBacnet(device: FieldDevice)}
        <Table.Row class="bg-muted/30 hover:bg-muted/40">
          <Table.Cell colspan={columnCount} class="p-0">
            {#if tableState.isBacnetLoading(device.id)}
              <div class="p-4">
                <Skeleton class="h-10 w-full" />
              </div>
            {:else}
              <BacnetObjectsEditor
                bacnetObjects={device.bacnet_objects ?? []}
                pendingEdits={tableState.editing.getBacnetPendingEdits(device.id) ?? new Map()}
                fieldErrors={tableState.editing.getBacnetFieldErrors(device.id) ?? new Map()}
                clientErrors={tableState.editing.getBacnetClientErrors(device.id) ?? new Map()}
                sharedEditors={tableState.getEditorsForDevice(device.id)}
                disabled={!tableState.canUpdateFieldDeviceBacnetObjects()}
                onEdit={(objectId, field, value) => {
                  tableState.editing.queueBacnetEdit(device.id, objectId, field, value);
                }}
                onUndoField={(objectId, field) => {
                  tableState.editing.discardBacnetObjectFieldEdit(device.id, objectId, field);
                }}
                onUndoRow={(objectId) => {
                  tableState.editing.discardBacnetObjectEdits(device.id, objectId);
                }}
              />
            {/if}
          </Table.Cell>
        </Table.Row>
      {/snippet}

      {#snippet renderGroup(group: TableGroupNode<FieldDevice, FieldDeviceGroupKey>)}
        <Table.Row class="bg-muted/35 hover:bg-muted/50">
          <Table.Cell colspan={columnCount} class="p-0">
            <Button
              type="button"
              variant="ghost"
              class="flex h-auto w-full justify-start gap-2 px-3 py-2 text-left font-medium"
              style={`padding-left: ${0.75 + group.level * 1.25}rem`}
              aria-expanded={tableState.view.grouping.isGroupExpanded(group.id)}
              onclick={() => tableState.view.grouping.toggleGroupExpansion(group.id)}
            >
              {#if tableState.view.grouping.isGroupExpanded(group.id)}
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
      {/snippet}

      {#if tableState.loading}
        <Table.LoadingRows loading {columnCount} rowCount={8} delayMs={0} />
      {:else if tableState.items.length === 0}
        <Table.Row>
          <Table.Cell colspan={columnCount} class="h-24 text-center">
            <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
              <p class="font-medium">{$t('field_device.empty.title')}</p>
              {#if tableState.searchText}
                <p class="text-sm">{$t('field_device.empty.search_hint')}</p>
              {/if}
            </div>
          </Table.Cell>
        </Table.Row>
      {:else if tableState.view.grouping.isGrouped}
        {#if renderWindow.topSpacerHeight > 0}
          <Table.Row>
            <Table.Cell colspan={columnCount} class="h-0 border-0 p-0">
              <div style={`height:${renderWindow.topSpacerHeight}px`} class="w-full"></div>
            </Table.Cell>
          </Table.Row>
        {/if}

        {#each renderWindow.rows as row (row.key)}
          {#if row.kind === 'group'}
            {@render renderGroup(row.group)}
          {:else if row.kind === 'device'}
            {@render renderDevice(row)}
          {:else if row.kind === 'bacnet'}
            {@render renderBacnet(row.device)}
          {/if}
        {/each}

        {#if renderWindow.bottomSpacerHeight > 0}
          <Table.Row>
            <Table.Cell colspan={columnCount} class="h-0 border-0 p-0">
              <div style={`height:${renderWindow.bottomSpacerHeight}px`} class="w-full"></div>
            </Table.Cell>
          </Table.Row>
        {/if}
      {:else}
        {#if renderWindow.topSpacerHeight > 0}
          <Table.Row>
            <Table.Cell colspan={columnCount} class="h-0 border-0 p-0">
              <div style={`height:${renderWindow.topSpacerHeight}px`} class="w-full"></div>
            </Table.Cell>
          </Table.Row>
        {/if}

        {#each renderWindow.rows as row (row.key)}
          {#if row.kind === 'device'}
            {@render renderDevice(row)}
          {:else if row.kind === 'bacnet'}
            {@render renderBacnet(row.device)}
          {/if}
        {/each}

        {#if renderWindow.bottomSpacerHeight > 0}
          <Table.Row>
            <Table.Cell colspan={columnCount} class="h-0 border-0 p-0">
              <div style={`height:${renderWindow.bottomSpacerHeight}px`} class="w-full"></div>
            </Table.Cell>
          </Table.Row>
        {/if}
      {/if}
    </Table.Body>
  </Table.Root>
</div>
