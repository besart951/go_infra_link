<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { keyboardTableCell } from '$lib/actions/keyboardTableNavigation.js';
  import { ChevronDown, ChevronRight } from '@lucide/svelte';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import TableApparatSelect from '../table-selects/TableApparatSelect.svelte';
  import TableSystemPartSelect from '../table-selects/TableSystemPartSelect.svelte';
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

  function toDisplayString(value: unknown, isNumeric = false): string {
    if (value === null || value === undefined || value === '') return '';
    if (isNumeric && typeof value === 'number') return String(value);
    return String(value);
  }

  function handleApparatChange(newApparatId: string) {
    if (!newApparatId || newApparatId === device.apparat_id) return;
    rowState.editing.queueEdit(device.id, 'apparat_id', newApparatId);
  }

  function handleSystemPartChange(newSystemPartId: string) {
    if (!newSystemPartId || newSystemPartId === device.system_part_id) return;
    rowState.editing.queueEdit(device.id, 'system_part_id', newSystemPartId);
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

  function editCell(column: string): Record<string, string> {
    return keyboardTableCell(device.id, column, { activate: 'edit' });
  }

  function focusCell(column: string): Record<string, string> {
    return keyboardTableCell(device.id, column, { activate: 'focus' });
  }
</script>

<Table.Row
  class={[rowState.loading ? 'opacity-60' : '', rowState.isSelected(device.id) ? 'bg-muted/50' : '']
    .filter(Boolean)
    .join(' ')}
>
  <Table.Cell class="p-2">
    <Checkbox
      checked={rowState.isSelected(device.id)}
      onCheckedChange={() => rowState.toggleSelection(device.id)}
      aria-label={$t('field_device.table.select_aria', { label: device.bmk || device.id })}
    />
  </Table.Cell>
  <Table.Cell class="p-2">
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
  <Table.Cell class="font-medium">
    {formatFieldDeviceSPSControllerSystemType(device)}
  </Table.Cell>
  <Table.Cell class="p-1">
    <div
      class={getEditingFieldClass('bmk')}
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
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'bmk', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="max-w-48 p-1">
    <div
      class={getEditingFieldClass('description')}
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
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'description', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="p-1">
    <div
      class={getEditingFieldClass('text_fix')}
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
        onSave={(value) => {
          rowState.editing.queueEdit(device.id, 'text_fix', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="p-1">
    <div
      class={getEditingFieldClass('apparat_nr')}
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
        onSave={(value) => {
          rowState.editing.queueEdit(
            device.id,
            'apparat_nr',
            value ? parseInt(value, 10) : undefined
          );
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell
    class={getEditingFieldClass('apparat_id')}
    title={getFieldPreviewTitle('apparat_id')}
    {...focusCell('apparat_id')}
  >
    <TableApparatSelect
      items={rowState.allApparats}
      value={device.apparat_id}
      width="w-full"
      disabled={!rowState.canUpdateFieldDevice()}
      error={rowState.editing.getFieldError(device.id, 'apparat_id')}
      onValueChange={handleApparatChange}
    />
  </Table.Cell>
  <Table.Cell
    class={getEditingFieldClass('system_part_id')}
    title={getFieldPreviewTitle('system_part_id')}
    {...focusCell('system_part_id')}
  >
    <TableSystemPartSelect
      items={rowState.allSystemParts}
      value={device.system_part_id || ''}
      width="w-full"
      disabled={!rowState.canUpdateFieldDevice()}
      error={rowState.editing.getFieldError(device.id, 'system_part_id')}
      onValueChange={handleSystemPartChange}
    />
  </Table.Cell>
  <Table.Cell class="text-center">
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
      class={`text-xs ${getEditingFieldClass('specification.specification_supplier')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_supplier',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.specification_brand')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_brand',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.specification_type')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'specification_type',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.additional_info_motor_valve')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_info_motor_valve',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.additional_info_size')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_info_size',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.additional_information_installation_location')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'additional_information_installation_location',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.electrical_connection_ph')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_ph',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.electrical_connection_acdc')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_acdc',
            value === '' ? null : value
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.electrical_connection_amperage')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_amperage',
            value === '' ? null : value ? parseFloat(value) : null
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.electrical_connection_power')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_power',
            value === '' ? null : value ? parseFloat(value) : null
          );
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.electrical_connection_rotation')}`}
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
        onSave={(value) => {
          rowState.editing.queueSpecEdit(
            device.id,
            'electrical_connection_rotation',
            value === '' ? null : value ? parseInt(value, 10) : null
          );
        }}
      />
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
