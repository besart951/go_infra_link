<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { EditableCell } from '$lib/components/ui/editable-cell/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import { ChevronDown, ChevronRight } from '@lucide/svelte';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import TableApparatSelect from '../table-selects/TableApparatSelect.svelte';
  import TableSystemPartSelect from '../table-selects/TableSystemPartSelect.svelte';
  import type { FieldDevice } from '$lib/domain/facility/index.js';
  import type { SharedFieldDeviceEditor } from '$lib/services/projectCollaboration.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  interface Props {
    device: FieldDevice;
  }

  let { device }: Props = $props();

  const t = createTranslator();
  const state = useFieldDeviceState();

  function toDisplayString(value: unknown, isNumeric = false): string {
    if (value === null || value === undefined || value === '') return '';
    if (isNumeric && typeof value === 'number') return String(value);
    return String(value);
  }

  function formatSPSControllerSystemType(fieldDevice: FieldDevice): string {
    const systemType = fieldDevice.sps_controller_system_type;
    if (!systemType) return '-';

    const deviceName = systemType.sps_controller_name ?? '';
    const number =
      systemType.number === null || systemType.number === undefined
        ? ''
        : String(systemType.number).padStart(4, '0');
    const documentName = systemType.document_name ?? '';

    let systemTypePart = '';
    if (number && documentName) {
      systemTypePart = `${number} - ${documentName}`;
    } else if (number) {
      systemTypePart = number;
    } else if (documentName) {
      systemTypePart = documentName;
    }

    if (deviceName && systemTypePart) return `${deviceName}_${systemTypePart}`;
    if (deviceName) return deviceName;
    if (systemTypePart) return systemTypePart;
    return '-';
  }

  function handleApparatChange(newApparatId: string) {
    if (!newApparatId || newApparatId === device.apparat_id) return;
    state.editing.queueEdit(device.id, 'apparat_id', newApparatId);
  }

  function handleSystemPartChange(newSystemPartId: string) {
    if (!newSystemPartId || newSystemPartId === device.system_part_id) return;
    state.editing.queueEdit(device.id, 'system_part_id', newSystemPartId);
  }

  const hasBacnetErrors = $derived.by(
    () =>
      state.editing.getBacnetFieldErrors(device.id).size > 0 ||
      state.editing.getBacnetClientErrors(device.id).size > 0
  );
  const collaborators = $derived(state.getEditorsForDevice(device.id));

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
      ? 'bg-yellow-100/40 dark:bg-yellow-900/20 cursor-help'
      : '';
  }
</script>

<Table.Row
  class={[state.loading ? 'opacity-60' : '', state.isSelected(device.id) ? 'bg-muted/50' : '']
    .filter(Boolean)
    .join(' ')}
>
  <Table.Cell class="p-2">
    <Checkbox
      checked={state.isSelected(device.id)}
      onCheckedChange={() => state.toggleSelection(device.id)}
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
      onclick={() => state.toggleBacnetExpansion(device.id)}
      title={$t('field_device.table.bacnet_expand')}
    >
      {#if state.isBacnetExpanded(device.id)}
        <ChevronDown class="h-4 w-4" />
      {:else}
        <ChevronRight class="h-4 w-4" />
      {/if}
    </Button>
  </Table.Cell>
  <Table.Cell class="font-medium">
    {formatSPSControllerSystemType(device)}
  </Table.Cell>
  <Table.Cell class="p-1">
    <div class={getEditingFieldClass('bmk')} title={getFieldPreviewTitle('bmk')}>
      <EditableCell
        value={device.bmk ?? ''}
        pendingValue={state.editing.getPendingValue(device.id, 'bmk')}
        type="text"
        maxlength={10}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isFieldDirty(device.id, 'bmk')}
        error={state.editing.getFieldError(device.id, 'bmk')}
        onSave={(value) => {
          state.editing.queueEdit(device.id, 'bmk', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="max-w-48 p-1">
    <div class={getEditingFieldClass('description')} title={getFieldPreviewTitle('description')}>
      <EditableCell
        value={device.description ?? ''}
        pendingValue={state.editing.getPendingValue(device.id, 'description')}
        type="text"
        maxlength={250}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isFieldDirty(device.id, 'description')}
        error={state.editing.getFieldError(device.id, 'description')}
        onSave={(value) => {
          state.editing.queueEdit(device.id, 'description', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="p-1">
    <div class={getEditingFieldClass('text_fix')} title={getFieldPreviewTitle('text_fix')}>
      <EditableCell
        value={device.text_fix ?? ''}
        pendingValue={state.editing.getPendingValue(device.id, 'text_fix')}
        type="text"
        maxlength={250}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isFieldDirty(device.id, 'text_fix')}
        error={state.editing.getFieldError(device.id, 'text_fix')}
        onSave={(value) => {
          state.editing.queueEdit(device.id, 'text_fix', value === '' ? null : value);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class="p-1">
    <div class={getEditingFieldClass('apparat_nr')} title={getFieldPreviewTitle('apparat_nr')}>
      <EditableCell
        value={device.apparat_nr}
        pendingValue={state.editing.getPendingValue(device.id, 'apparat_nr')}
        type="number"
        min={1}
        max={99}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isFieldDirty(device.id, 'apparat_nr')}
        error={state.editing.getFieldError(device.id, 'apparat_nr')}
        onSave={(value) => {
          state.editing.queueEdit(device.id, 'apparat_nr', value ? parseInt(value, 10) : undefined);
        }}
      />
    </div>
  </Table.Cell>
  <Table.Cell class={getEditingFieldClass('apparat_id')} title={getFieldPreviewTitle('apparat_id')}>
    <TableApparatSelect
      items={state.allApparats}
      value={device.apparat_id}
      width="w-full"
      disabled={!canPerform('update', 'fielddevice')}
      error={state.editing.getFieldError(device.id, 'apparat_id')}
      onValueChange={handleApparatChange}
    />
  </Table.Cell>
  <Table.Cell class={getEditingFieldClass('system_part_id')} title={getFieldPreviewTitle('system_part_id')}>
    <TableSystemPartSelect
      items={state.allSystemParts}
      value={device.system_part_id || ''}
      width="w-full"
      disabled={!canPerform('update', 'fielddevice')}
      error={state.editing.getFieldError(device.id, 'system_part_id')}
      onValueChange={handleSystemPartChange}
    />
  </Table.Cell>
  <Table.Cell class="text-center">
    {#if device.specification}
      <span
        class="inline-block h-2 w-2 rounded-full bg-green-500"
        title={$t('field_device.table.spec_available')}
      ></span>
    {:else}
      <span
        class="inline-block h-2 w-2 rounded-full bg-gray-300"
        title={$t('field_device.table.spec_missing')}
      ></span>
    {/if}
  </Table.Cell>
  {#if state.showSpecifications}
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.specification_supplier')}`}
      title={getFieldPreviewTitle('specification.specification_supplier')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_supplier)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'specification_supplier')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'specification_supplier')}
        error={state.editing.getFieldError(device.id, 'specification_supplier')}
        maxlength={250}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_brand)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'specification_brand')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'specification_brand')}
        error={state.editing.getFieldError(device.id, 'specification_brand')}
        maxlength={250}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.specification_type)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'specification_type')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'specification_type')}
        error={state.editing.getFieldError(device.id, 'specification_type')}
        maxlength={250}
        onSave={(value) => {
          state.editing.queueSpecEdit(device.id, 'specification_type', value === '' ? null : value);
        }}
      />
    </Table.Cell>
    <Table.Cell
      class={`text-xs ${getEditingFieldClass('specification.additional_info_motor_valve')}`}
      title={getFieldPreviewTitle('specification.additional_info_motor_valve')}
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_info_motor_valve)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'additional_info_motor_valve')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'additional_info_motor_valve')}
        error={state.editing.getFieldError(device.id, 'additional_info_motor_valve')}
        maxlength={250}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_info_size, true)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'additional_info_size')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'additional_info_size')}
        error={state.editing.getFieldError(device.id, 'additional_info_size')}
        type="number"
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.additional_information_installation_location)}
        pendingValue={state.editing.getPendingSpecValue(
          device.id,
          'additional_information_installation_location'
        )}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(
          device.id,
          'additional_information_installation_location'
        )}
        error={state.editing.getFieldError(
          device.id,
          'additional_information_installation_location'
        )}
        maxlength={250}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_ph, true)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'electrical_connection_ph')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'electrical_connection_ph')}
        error={state.editing.getFieldError(device.id, 'electrical_connection_ph')}
        type="number"
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_acdc)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'electrical_connection_acdc')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'electrical_connection_acdc')}
        error={state.editing.getFieldError(device.id, 'electrical_connection_acdc')}
        maxlength={2}
        placeholder={$t('field_device.table.acdc_placeholder')}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_amperage, true)}
        pendingValue={state.editing.getPendingSpecValue(
          device.id,
          'electrical_connection_amperage'
        )}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'electrical_connection_amperage')}
        error={state.editing.getFieldError(device.id, 'electrical_connection_amperage')}
        type="number"
        placeholder={$t('field_device.table.amperage_placeholder')}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_power, true)}
        pendingValue={state.editing.getPendingSpecValue(device.id, 'electrical_connection_power')}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'electrical_connection_power')}
        error={state.editing.getFieldError(device.id, 'electrical_connection_power')}
        type="number"
        placeholder={$t('field_device.table.power_placeholder')}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
    >
      <EditableCell
        value={toDisplayString(device.specification?.electrical_connection_rotation, true)}
        pendingValue={state.editing.getPendingSpecValue(
          device.id,
          'electrical_connection_rotation'
        )}
        disabled={!canPerform('update', 'fielddevice')}
        isDirty={state.editing.isSpecFieldDirty(device.id, 'electrical_connection_rotation')}
        error={state.editing.getFieldError(device.id, 'electrical_connection_rotation')}
        type="number"
        placeholder={$t('field_device.table.rotation_placeholder')}
        onSave={(value) => {
          state.editing.queueSpecEdit(
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
            void state.copyToClipboard(
              device.bmk?.trim() || (device.apparat_nr ? String(device.apparat_nr) : device.id)
            )}
        >
          {$t('facility.copy')}
        </DropdownMenu.Item>
        {#if canPerform('delete', 'fielddevice')}
          <DropdownMenu.Separator />
          <DropdownMenu.Item variant="destructive" onclick={() => void state.deleteDevice(device)}>
            {$t('common.delete')}
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </Table.Cell>
</Table.Row>
