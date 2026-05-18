<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { keyboardTableCell } from '$lib/actions/keyboardTableNavigation.js';
  import { ChevronDown, ChevronRight, RotateCcw, Undo2 } from '@lucide/svelte';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import TableApparatSelect from '../table-selects/TableApparatSelect.svelte';
  import TableSystemPartSelect from '../table-selects/TableSystemPartSelect.svelte';
  import {
    filterApparatsForRelationSource,
    filterSystemPartsForRelationSource,
    isSystemPartAllowedForApparat,
    mergeSelectedRelationOption
  } from '../table-selects/relationSelectOptions.js';
  import type { RelationFilterSource } from '../table-selects/relationSelectOptions.js';
  import { InlineUndoButton } from '$lib/components/ui/editable-cell/index.js';
  import type { FieldDevice } from '$lib/domain/facility/index.js';
  import type { SharedFieldDeviceEditor } from '$lib/services/projectCollaboration.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';
  import { formatFieldDeviceSPSControllerSystemType } from './state/FieldDeviceTableView.svelte.js';

  interface Props {
    device: FieldDevice;
  }

  let { device }: Props = $props();

  const t = createTranslator();
  const rowState = useFieldDeviceState();
  let historyOpen = $state(false);
  let relationFilterSource = $state<RelationFilterSource>(null);

  function toDisplayString(value: unknown, isNumeric = false): string {
    if (value === null || value === undefined || value === '') return '';
    if (isNumeric && typeof value === 'number') return String(value);
    return String(value);
  }

  function queueRelationEdit(
    field: 'apparat_id' | 'system_part_id',
    nextValue: string,
    originalValue: string
  ) {
    if (nextValue && nextValue === originalValue) {
      rowState.editing.discardFieldEdit(device.id, field);
      return;
    }

    rowState.editing.queueEdit(device.id, field, nextValue);
  }

  function handleApparatChange(newApparatId: string) {
    const nextApparatId = newApparatId || '';
    relationFilterSource = nextApparatId ? 'apparat_id' : null;
    queueRelationEdit('apparat_id', nextApparatId, device.apparat_id);

    if (
      nextApparatId &&
      systemPartSelectValue &&
      !isSystemPartAllowedForApparat(rowState.allApparats, nextApparatId, systemPartSelectValue)
    ) {
      rowState.editing.queueEdit(device.id, 'system_part_id', '');
    }
  }

  function handleSystemPartChange(newSystemPartId: string) {
    const nextSystemPartId = newSystemPartId || '';
    relationFilterSource = nextSystemPartId ? 'system_part_id' : null;
    queueRelationEdit('system_part_id', nextSystemPartId, device.system_part_id ?? '');

    if (
      nextSystemPartId &&
      apparatSelectValue &&
      !isSystemPartAllowedForApparat(rowState.allApparats, apparatSelectValue, nextSystemPartId)
    ) {
      rowState.editing.queueEdit(device.id, 'apparat_id', '');
    }
  }

  const hasBacnetErrors = $derived.by(
    () =>
      rowState.editing.getBacnetFieldErrors(device.id).size > 0 ||
      rowState.editing.getBacnetClientErrors(device.id).size > 0
  );
  const collaborators = $derived(rowState.getEditorsForDevice(device.id));

  function getEditorsForField(fieldName: string): SharedFieldDeviceEditor[] {
    return collaborators.filter(
      (collab: SharedFieldDeviceEditor) =>
        collab.changedFields && collab.changedFields.includes(fieldName)
    );
  }

  function getFieldPreviewTitle(fieldName: string): string | undefined {
    const editors = getEditorsForField(fieldName);
    if (editors.length === 0) return undefined;

    const lines = editors.map((editor) => {
      const value = editor.fieldValues?.[fieldName];
      let displayValue = '(empty)';
      if (value !== null && value !== undefined) {
        displayValue = typeof value === 'object' ? JSON.stringify(value) : String(value);
      }
      return `${editor.firstName} ${editor.lastName}: ${displayValue}`;
    });

    return lines.join('\n');
  }

  function getEditingFieldClass(fieldName: string): string {
    return getEditorsForField(fieldName).length > 0
      ? 'bg-warning-muted/60 dark:bg-warning-muted/60 cursor-help'
      : '';
  }

  function cellClass(widthClass: string, extra = '', fieldName?: string): string {
    return [
      widthClass,
      'content-width-cell overflow-hidden',
      extra,
      fieldName ? getEditingFieldClass(fieldName) : ''
    ]
      .filter(Boolean)
      .join(' ');
  }

  function editingWrapperClass(fieldName: string): string {
    return ['min-w-0', getEditingFieldClass(fieldName)].filter(Boolean).join(' ');
  }

  function editCell(column: string): Record<string, string> {
    return keyboardTableCell(device.id, column, { activate: 'edit' });
  }

  function focusCell(column: string): Record<string, string> {
    return keyboardTableCell(device.id, column, { activate: 'focus' });
  }

  const spsControllerSystemTypeLabel = $derived(formatFieldDeviceSPSControllerSystemType(device));
  const apparatSelectValue = $derived(
    rowState.editing.getPendingValue(device.id, 'apparat_id') ?? device.apparat_id
  );
  const systemPartSelectValue = $derived(
    rowState.editing.getPendingValue(device.id, 'system_part_id') ?? device.system_part_id ?? ''
  );
  const selectedApparatFallback = $derived(
    apparatSelectValue === device.apparat_id ? device.apparat : undefined
  );
  const selectedSystemPartFallback = $derived(
    systemPartSelectValue === (device.system_part_id ?? '') ? device.system_part : undefined
  );
  const apparatSelectItems = $derived(
    mergeSelectedRelationOption(
      filterApparatsForRelationSource(
        rowState.allApparats,
        systemPartSelectValue,
        relationFilterSource
      ),
      selectedApparatFallback
    )
  );
  const systemPartSelectItems = $derived(
    mergeSelectedRelationOption(
      filterSystemPartsForRelationSource(
        rowState.allSystemParts,
        rowState.allApparats,
        apparatSelectValue,
        relationFilterSource
      ),
      selectedSystemPartFallback
    )
  );
  const apparatSelectDirty = $derived(rowState.editing.isFieldDirty(device.id, 'apparat_id'));
  const systemPartSelectDirty = $derived(
    rowState.editing.isFieldDirty(device.id, 'system_part_id')
  );
  const apparatNrSuggestion = $derived(
    rowState.editing.getFieldSuggestion(device.id, 'apparat_nr', rowState.items)
  );
  const hasFieldDevicePendingEdits = $derived(
    rowState.editing.hasPendingFieldDeviceEditsForDevice(device.id)
  );
  const hasDevicePendingEdits = $derived(rowState.editing.hasPendingEditsForDevice(device.id));

  const undoFieldTitle = $derived($t('field_device.editing.undo_field'));
  const undoFieldDeviceTitle = $derived($t('field_device.editing.undo_field_device'));
  const undoDeviceTitle = $derived($t('field_device.editing.undo_device'));
</script>

<Table.Row
  class={[
    'group/fd-row',
    rowState.loading ? 'opacity-60' : '',
    rowState.isSelected(device.id) ? 'bg-muted/50' : ''
  ]
    .filter(Boolean)
    .join(' ')}
>
  <Table.Cell class="w-8 max-w-8 min-w-8 px-0! py-1!">
    <div class="flex justify-center">
      <Checkbox
        checked={rowState.isSelected(device.id)}
        onCheckedChange={() => rowState.toggleSelection(device.id)}
        aria-label={$t('field_device.table.select_aria', { label: device.bmk || device.id })}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="w-6 max-w-6 min-w-6 px-0! py-1!">
    <Button
      variant="ghost"
      size="sm"
      class={[
        'h-6 w-6 p-0',
        hasBacnetErrors ? 'text-destructive ring-1 ring-destructive/40 hover:text-destructive' : ''
      ]
        .filter(Boolean)
        .join(' ')}
      onclick={() => void rowState.toggleBacnetExpansion(device.id)}
      title={$t('field_device.table.bacnet_expand')}
    >
      {#if rowState.isBacnetExpanded(device.id)}
        <ChevronDown class="h-4 w-4" />
      {:else}
        <ChevronRight class="h-4 w-4" />
      {/if}
    </Button>
  </Table.Cell>
  <Table.Cell class="w-fit max-w-96 overflow-hidden text-xs font-medium">
    {#if spsControllerSystemTypeLabel && spsControllerSystemTypeLabel !== '-'}
      <Tooltip.Provider>
        <Tooltip.Root>
          <Tooltip.Trigger>
            {#snippet child({ props })}
              <span {...props} class="inline-block max-w-full truncate">
                {spsControllerSystemTypeLabel}
              </span>
            {/snippet}
          </Tooltip.Trigger>
          <Tooltip.Content side="top" class="max-w-xs">
            <p class="wrap-break-word">{spsControllerSystemTypeLabel}</p>
          </Tooltip.Content>
        </Tooltip.Root>
      </Tooltip.Provider>
    {:else}
      <span class="inline-block max-w-full truncate">
        {spsControllerSystemTypeLabel}
      </span>
    {/if}
  </Table.Cell>
  <Table.Cell class="content-width-cell w-fit max-w-36 overflow-hidden p-1">
    <div
      class={editingWrapperClass('bmk')}
      title={getFieldPreviewTitle('bmk')}
      {...editCell('bmk')}
    >
      <EditableCell
        value={device.bmk ?? ''}
        pendingValue={rowState.editing.getPendingValue(device.id, 'bmk')}
        type="text"
        maxlength={10}
        disabled={!rowState.canUpdateFieldDevice()}
        isDirty={rowState.editing.isFieldDirty(device.id, 'bmk')}
        error={rowState.editing.getFieldError(device.id, 'bmk')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'bmk', value === '' ? null : value);
        }}
        onUndo={() => rowState.editing.discardFieldEdit(device.id, 'bmk')}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="content-width-cell w-fit max-w-80 overflow-hidden p-1">
    <div
      class={editingWrapperClass('description')}
      title={getFieldPreviewTitle('description')}
      {...editCell('description')}
    >
      <EditableCell
        value={device.description ?? ''}
        pendingValue={rowState.editing.getPendingValue(device.id, 'description')}
        type="text"
        maxlength={250}
        disabled={!rowState.canUpdateFieldDevice()}
        isDirty={rowState.editing.isFieldDirty(device.id, 'description')}
        error={rowState.editing.getFieldError(device.id, 'description')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'description', value === '' ? null : value);
        }}
        onUndo={() => rowState.editing.discardFieldEdit(device.id, 'description')}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="content-width-cell w-fit max-w-80 overflow-hidden p-1">
    <div
      class={editingWrapperClass('text_fix')}
      title={getFieldPreviewTitle('text_fix')}
      {...editCell('text_fix')}
    >
      <EditableCell
        value={device.text_fix ?? ''}
        pendingValue={rowState.editing.getPendingValue(device.id, 'text_fix')}
        type="text"
        maxlength={250}
        disabled={!rowState.canUpdateFieldDevice()}
        isDirty={rowState.editing.isFieldDirty(device.id, 'text_fix')}
        error={rowState.editing.getFieldError(device.id, 'text_fix')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'text_fix', value === '' ? null : value);
        }}
        onUndo={() => rowState.editing.discardFieldEdit(device.id, 'text_fix')}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="content-width-cell w-fit max-w-20 overflow-hidden p-1">
    <div
      class={editingWrapperClass('apparat_nr')}
      title={getFieldPreviewTitle('apparat_nr')}
      {...editCell('apparat_nr')}
    >
      <EditableCell
        value={device.apparat_nr}
        pendingValue={rowState.editing.getPendingValue(device.id, 'apparat_nr')}
        type="number"
        min={1}
        max={99}
        disabled={!rowState.canUpdateFieldDevice()}
        isDirty={rowState.editing.isFieldDirty(device.id, 'apparat_nr')}
        error={rowState.editing.getFieldError(device.id, 'apparat_nr')}
        suggestion={apparatNrSuggestion !== undefined ? String(apparatNrSuggestion) : undefined}
        suggestionLabel={apparatNrSuggestion !== undefined
          ? $t('field_device.editing.apparat_nr_lowest_available', {
              value: apparatNrSuggestion
            })
          : undefined}
        suggestionActionLabel={apparatNrSuggestion !== undefined
          ? $t('field_device.editing.apparat_nr_use_available_short', {
              value: apparatNrSuggestion
            })
          : undefined}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueEdit(
            device.id,
            'apparat_nr',
            value ? parseInt(value, 10) : undefined
          );
        }}
        onApplySuggestion={(value) => {
          rowState.editing.queueEdit(device.id, 'apparat_nr', parseInt(value, 10));
        }}
        onUndo={() => rowState.editing.discardFieldEdit(device.id, 'apparat_nr')}
      />
    </div>
  </Table.Cell>
  <Table.Cell
    class={cellClass('w-fit max-w-40 min-w-18', 'p-1', 'apparat_id')}
    title={getFieldPreviewTitle('apparat_id')}
    {...focusCell('apparat_id')}
  >
    <div
      class={['group/undo relative rounded-md', apparatSelectDirty ? 'ring-1 ring-ring' : '']
        .filter(Boolean)
        .join(' ')}
    >
      <TableApparatSelect
        items={apparatSelectItems}
        value={apparatSelectValue}
        width="w-fit max-w-40 min-w-[4.5rem]"
        popupWidth="w-52"
        disabled={!rowState.canUpdateFieldDevice()}
        error={rowState.editing.getFieldError(device.id, 'apparat_id')}
        onValueChange={handleApparatChange}
        clearable
      />
      {#if apparatSelectDirty}
        <InlineUndoButton
          title={undoFieldTitle}
          onclick={() => {
            rowState.editing.discardFieldEdit(device.id, 'apparat_id');
            relationFilterSource = null;
          }}
        />
      {/if}
    </div>
  </Table.Cell>
  <Table.Cell
    class={cellClass('w-fit max-w-40 min-w-18', 'p-1', 'system_part_id')}
    title={getFieldPreviewTitle('system_part_id')}
    {...focusCell('system_part_id')}
  >
    <div
      class={['group/undo relative rounded-md', systemPartSelectDirty ? 'ring-1 ring-ring' : '']
        .filter(Boolean)
        .join(' ')}
    >
      <TableSystemPartSelect
        items={systemPartSelectItems}
        value={systemPartSelectValue}
        width="w-fit max-w-40 min-w-[4.5rem]"
        popupWidth="w-52"
        disabled={!rowState.canUpdateFieldDevice()}
        error={rowState.editing.getFieldError(device.id, 'system_part_id')}
        onValueChange={handleSystemPartChange}
        clearable
      />
      {#if systemPartSelectDirty}
        <InlineUndoButton
          title={undoFieldTitle}
          onclick={() => {
            rowState.editing.discardFieldEdit(device.id, 'system_part_id');
            relationFilterSource = null;
          }}
        />
      {/if}
    </div>
  </Table.Cell>
  <Table.Cell class="w-10 max-w-10 min-w-10 text-center">
    {#if device.specification_id || device.specification}
      <span
        class="inline-block h-2 w-2 rounded-full bg-success"
        title={$t('field_device.table.spec_available')}
      ></span>
    {:else}
      <span
        class="inline-block h-2 w-2 rounded-full bg-muted-foreground/40"
        title={$t('field_device.table.spec_missing')}
      ></span>
    {/if}
  </Table.Cell>
  {#if rowState.showSpecifications}
    <Table.Cell
      class={cellClass('w-fit max-w-64', 'p-1 text-xs', 'specification.specification_supplier')}
      title={getFieldPreviewTitle('specification.specification_supplier')}
      {...editCell('specification_supplier')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_supplier)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'specification_supplier')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'specification_supplier')}
        error={rowState.editing.getFieldError(device.id, 'specification_supplier')}
        maxlength={250}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_supplier',
            value === '' ? null : value
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'specification_supplier')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass('w-fit max-w-64', 'p-1 text-xs', 'specification.specification_brand')}
      title={getFieldPreviewTitle('specification.specification_brand')}
      {...editCell('specification_brand')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_brand)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'specification_brand')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'specification_brand')}
        error={rowState.editing.getFieldError(device.id, 'specification_brand')}
        maxlength={250}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_brand',
            value === '' ? null : value
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'specification_brand')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass('w-fit max-w-64', 'p-1 text-xs', 'specification.specification_type')}
      title={getFieldPreviewTitle('specification.specification_type')}
      {...editCell('specification_type')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_type)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'specification_type')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'specification_type')}
        error={rowState.editing.getFieldError(device.id, 'specification_type')}
        maxlength={250}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_type',
            value === '' ? null : value
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'specification_type')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass(
        'w-fit max-w-64',
        'p-1 text-xs',
        'specification.additional_info_motor_valve'
      )}
      title={getFieldPreviewTitle('specification.additional_info_motor_valve')}
      {...editCell('additional_info_motor_valve')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_info_motor_valve)}
        pendingValue={rowState.editing.getPendingSpecValue(
          device.id,
          'additional_info_motor_valve'
        )}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'additional_info_motor_valve')}
        error={rowState.editing.getFieldError(device.id, 'additional_info_motor_valve')}
        maxlength={250}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_info_motor_valve',
            value === '' ? null : value
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'additional_info_motor_valve')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass('w-fit max-w-24', 'p-1 text-xs', 'specification.additional_info_size')}
      title={getFieldPreviewTitle('specification.additional_info_size')}
      {...editCell('additional_info_size')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_info_size, true)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'additional_info_size')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'additional_info_size')}
        error={rowState.editing.getFieldError(device.id, 'additional_info_size')}
        type="number"
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_info_size',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'additional_info_size')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass(
        'w-fit max-w-80',
        'p-1 text-xs',
        'specification.additional_information_installation_location'
      )}
      title={getFieldPreviewTitle('specification.additional_information_installation_location')}
      {...editCell('additional_information_installation_location')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_information_installation_location)}
        pendingValue={rowState.editing.getPendingSpecValue(
          device.id,
          'additional_information_installation_location'
        )}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(
          device.id,
          'additional_information_installation_location'
        )}
        error={rowState.editing.getFieldError(
          device.id,
          'additional_information_installation_location'
        )}
        maxlength={250}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_information_installation_location',
            value === '' ? null : value
          );
        }}
        onUndo={() =>
          rowState.editing.discardSpecEdit(
            device.id,
            'additional_information_installation_location'
          )}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass('w-fit max-w-24', 'p-1 text-xs', 'specification.electrical_connection_ph')}
      title={getFieldPreviewTitle('specification.electrical_connection_ph')}
      {...editCell('electrical_connection_ph')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_ph, true)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'electrical_connection_ph')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'electrical_connection_ph')}
        error={rowState.editing.getFieldError(device.id, 'electrical_connection_ph')}
        type="number"
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_ph',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'electrical_connection_ph')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass('w-fit max-w-24', 'p-1 text-xs', 'specification.electrical_connection_acdc')}
      title={getFieldPreviewTitle('specification.electrical_connection_acdc')}
      {...editCell('electrical_connection_acdc')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_acdc)}
        pendingValue={rowState.editing.getPendingSpecValue(device.id, 'electrical_connection_acdc')}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'electrical_connection_acdc')}
        error={rowState.editing.getFieldError(device.id, 'electrical_connection_acdc')}
        maxlength={2}
        placeholder={$t('field_device.table.acdc_placeholder')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_acdc',
            value === '' ? null : value
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'electrical_connection_acdc')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass(
        'w-fit max-w-32',
        'p-1 text-xs',
        'specification.electrical_connection_amperage'
      )}
      title={getFieldPreviewTitle('specification.electrical_connection_amperage')}
      {...editCell('electrical_connection_amperage')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_amperage, true)}
        pendingValue={rowState.editing.getPendingSpecValue(
          device.id,
          'electrical_connection_amperage'
        )}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'electrical_connection_amperage')}
        error={rowState.editing.getFieldError(device.id, 'electrical_connection_amperage')}
        type="number"
        placeholder={$t('field_device.table.amperage_placeholder')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_amperage',
            value === '' ? null : value ? parseFloat(value) : null
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'electrical_connection_amperage')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass(
        'w-fit max-w-32',
        'p-1 text-xs',
        'specification.electrical_connection_power'
      )}
      title={getFieldPreviewTitle('specification.electrical_connection_power')}
      {...editCell('electrical_connection_power')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_power, true)}
        pendingValue={rowState.editing.getPendingSpecValue(
          device.id,
          'electrical_connection_power'
        )}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'electrical_connection_power')}
        error={rowState.editing.getFieldError(device.id, 'electrical_connection_power')}
        type="number"
        placeholder={$t('field_device.table.power_placeholder')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_power',
            value === '' ? null : value ? parseFloat(value) : null
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'electrical_connection_power')}
      />
    </Table.Cell>
    <Table.Cell
      class={cellClass(
        'w-fit max-w-32',
        'p-1 text-xs',
        'specification.electrical_connection_rotation'
      )}
      title={getFieldPreviewTitle('specification.electrical_connection_rotation')}
      {...editCell('electrical_connection_rotation')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_rotation, true)}
        pendingValue={rowState.editing.getPendingSpecValue(
          device.id,
          'electrical_connection_rotation'
        )}
        disabled={!rowState.canUpdateFieldDeviceSpecification()}
        isDirty={rowState.editing.isSpecFieldDirty(device.id, 'electrical_connection_rotation')}
        error={rowState.editing.getFieldError(device.id, 'electrical_connection_rotation')}
        type="number"
        placeholder={$t('field_device.table.rotation_placeholder')}
        undoTitle={undoFieldTitle}
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_rotation',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
        onUndo={() => rowState.editing.discardSpecEdit(device.id, 'electrical_connection_rotation')}
      />
    </Table.Cell>
  {/if}
  <Table.Cell class="w-20 max-w-20 min-w-20 text-right">
    <div class="flex items-center justify-end gap-1">
      {#if hasFieldDevicePendingEdits}
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-7 w-7 opacity-0 transition-opacity group-hover/fd-row:opacity-100 focus-visible:opacity-100"
          title={undoFieldDeviceTitle}
          onclick={() => {
            rowState.editing.discardDeviceFieldEdits(device.id);
            relationFilterSource = null;
          }}
        >
          <Undo2 class="size-4" />
        </Button>
      {/if}
      {#if hasDevicePendingEdits}
        <Button
          type="button"
          variant="ghost"
          size="icon"
          class="h-7 w-7 opacity-0 transition-opacity group-hover/fd-row:opacity-100 focus-visible:opacity-100"
          title={undoDeviceTitle}
          onclick={() => {
            rowState.editing.discardDeviceEdits(device.id);
            relationFilterSource = null;
          }}
        >
          <RotateCcw class="size-4" />
        </Button>
      {/if}
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
            onclick={() =>
              void rowState.copyToClipboard(
                device.bmk?.trim() || (device.apparat_nr ? String(device.apparat_nr) : device.id)
              )}
          >
            {$t('facility.copy')}
          </DropdownMenu.Item>
          <DropdownMenu.Item onclick={() => (historyOpen = true)}>
            {$t('history.open')}
          </DropdownMenu.Item>
          {#if rowState.canDeleteFieldDevice()}
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              variant="destructive"
              onclick={() => void rowState.deleteDevice(device)}
            >
              {$t('common.delete')}
            </DropdownMenu.Item>
          {/if}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </div>
  </Table.Cell>
</Table.Row>

<HistoryTimelineDialog
  bind:open={historyOpen}
  title={`${$t('history.title')}: ${device.bmk ?? device.id}`}
  scopeType="field_device"
  scopeId={device.id}
  projectId={rowState.effectiveProjectId}
  onRestored={() => rowState.reload()}
/>

<style>
  .content-width-cell :global(button:not(.editable-cell-display)) {
    width: fit-content;
    max-width: 100%;
  }

  .content-width-cell :global(.editable-cell-display),
  .content-width-cell :global(.editable-cell-editor) {
    width: 100%;
    max-width: 100%;
  }

  .content-width-cell :global(input) {
    min-width: 0;
  }

  .content-width-cell :global(span),
  .content-width-cell :global(code) {
    max-width: 100%;
  }
</style>
