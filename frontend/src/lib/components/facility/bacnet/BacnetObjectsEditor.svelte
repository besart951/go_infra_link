<script lang="ts">
  /**
   * BacnetObjectsEditor Component
   * Editable BACnet objects table for inline editing within expanded field device rows
   */
  import {
    EditableCell,
    EditableSelectCell,
    EditableBooleanCell,
    InlineUndoButton
  } from '$lib/components/ui/editable-cell/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { resizableTableColumns } from '$lib/actions/resizableTableColumns.js';
  import { keyboardTableCell } from '$lib/actions/keyboardTableNavigation.js';
  import BacnetAlarmValuesEditor from './BacnetAlarmValuesEditor.svelte';
  import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
  import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
  import { createCachedFetchById } from '$lib/infrastructure/api/createCachedFetchById.js';
  import {
    BACNET_SOFTWARE_TYPES,
    BACNET_HARDWARE_TYPES
  } from '$lib/domain/facility/bacnet-object.js';
  import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
  import type { BacnetObjectInput } from '$lib/domain/facility/field-device.js';
  import type { StateText, NotificationClass } from '$lib/domain/facility/index.js';
  import type { SharedFieldDeviceEditor } from '$lib/services/projectCollaboration.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { BellRing, Undo2 } from '@lucide/svelte';

  interface Props {
    bacnetObjects: BacnetObject[];
    pendingEdits: Map<string, Partial<BacnetObjectInput>>;
    fieldErrors: Map<string, Record<string, string>>;
    clientErrors: Map<string, Record<string, string>>;
    sharedEditors?: SharedFieldDeviceEditor[];
    disabled?: boolean;
    onEdit: (objectId: string, field: string, value: unknown) => void;
    onUndoField?: (objectId: string, field: string) => void;
    onUndoRow?: (objectId: string) => void;
  }

  let {
    bacnetObjects,
    pendingEdits,
    fieldErrors,
    clientErrors,
    sharedEditors = [],
    disabled = false,
    onEdit,
    onUndoField,
    onUndoRow
  }: Props = $props();

  const t = createTranslator();

  const softwareTypeOptions = BACNET_SOFTWARE_TYPES.map((t) => ({
    value: t.value,
    label: t.value.toUpperCase()
  }));
  const hardwareTypeOptions = BACNET_HARDWARE_TYPES.map((t) => ({
    value: t.value,
    label: t.value.toUpperCase()
  }));
  const fetchStateTextByIdCached = createCachedFetchById('state-text', (id) =>
    stateTextRepository.get(id)
  );
  const fetchNotificationClassByIdCached = createCachedFetchById('notification-class', (id) =>
    notificationClassRepository.get(id)
  );

  function isDirty(objectId: string, field: string): boolean {
    const edits = pendingEdits.get(objectId);
    return edits ? field in edits : false;
  }

  function hasObjectEdits(objectId: string): boolean {
    return Object.keys(pendingEdits.get(objectId) ?? {}).length > 0;
  }

  function getPendingTextValue(
    objectId: string,
    field: string,
    originalValue: string
  ): string | undefined {
    const edits = pendingEdits.get(objectId);
    if (!edits || !(field in edits)) return undefined;
    const val = (edits as Record<string, unknown>)[field];
    return val !== undefined ? String(val) : undefined;
  }

  function getPendingBoolValue(objectId: string, field: string): boolean | undefined {
    const edits = pendingEdits.get(objectId);
    if (!edits || !(field in edits)) return undefined;
    return (edits as Record<string, unknown>)[field] as boolean;
  }

  function getPendingIdValue(objectId: string, field: string, originalValue?: string): string {
    const edits = pendingEdits.get(objectId);
    if (!edits || !(field in edits)) {
      return originalValue ?? '';
    }
    const value = (edits as Record<string, unknown>)[field];
    return typeof value === 'string' ? value : '';
  }

  function hasTextIndividual(obj: BacnetObject): boolean {
    const edits = pendingEdits.get(obj.id);
    if (edits && 'text_individual' in edits) {
      return true;
    }
    return !!obj.text_individual;
  }

  function getFieldError(objectId: string, field: string): string | undefined {
    return fieldErrors.get(objectId)?.[field] || clientErrors.get(objectId)?.[field];
  }

  function getCollaborationFieldKey(objectId: string, field: string): string {
    return `bacnet_objects.${objectId}.${field}`;
  }

  function getEditorsForField(objectId: string, field: string): SharedFieldDeviceEditor[] {
    const collaborationField = getCollaborationFieldKey(objectId, field);
    return sharedEditors.filter((editor) => editor.changedFields.includes(collaborationField));
  }

  function getPreviewTitle(objectId: string, field: string): string | undefined {
    const collaborationField = getCollaborationFieldKey(objectId, field);
    const editors = getEditorsForField(objectId, field);
    if (editors.length === 0) return undefined;

    return editors
      .map((editor) => {
        const value = editor.fieldValues?.[collaborationField];
        const displayValue =
          value === null || value === undefined
            ? '(empty)'
            : typeof value === 'object'
              ? JSON.stringify(value)
              : String(value);
        return `${editor.firstName} ${editor.lastName}: ${displayValue}`;
      })
      .join('\n');
  }

  function getCollaborationClass(objectId: string, field: string): string {
    return getEditorsForField(objectId, field).length > 0
      ? 'rounded-md bg-warning-muted/60 dark:bg-warning-muted/60 cursor-help'
      : '';
  }

  function editCell(objectId: string, column: string): Record<string, string> {
    return keyboardTableCell(`bacnet:${objectId}`, `bacnet.${column}`, { activate: 'edit' });
  }

  function focusCell(objectId: string, column: string): Record<string, string> {
    return keyboardTableCell(`bacnet:${objectId}`, `bacnet.${column}`, { activate: 'focus' });
  }

  function hasAlarmType(obj: BacnetObject): boolean {
    const pendingAlarmTypeId = getPendingTextValue(
      obj.id,
      'alarm_type_id',
      obj.alarm_type_id || ''
    );
    const alarmTypeId = pendingAlarmTypeId ?? obj.alarm_type_id ?? '';
    return alarmTypeId.trim().length > 0;
  }

  async function fetchStateTexts(search: string): Promise<StateText[]> {
    const res = await stateTextRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    return res.items;
  }

  async function fetchStateTextById(id: string): Promise<StateText> {
    return fetchStateTextByIdCached(id) as Promise<StateText>;
  }

  function formatStateTextLabel(item: StateText): string {
    return String(item.ref_number);
  }

  function formatStateTextTooltip(item: StateText): string {
    const lines: string[] = [`#${item.ref_number}`];
    for (let index = 1; index <= 16; index++) {
      const key = `state_text${index}` as keyof StateText;
      const value = item[key];
      if (typeof value === 'string' && value.trim()) {
        lines.push(`${index}. ${value.trim()}`);
      }
    }
    return lines.join('\n');
  }

  async function fetchNotificationClasses(search: string): Promise<NotificationClass[]> {
    const res = await notificationClassRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    return res.items;
  }

  async function fetchNotificationClassById(id: string): Promise<NotificationClass> {
    return fetchNotificationClassByIdCached(id) as Promise<NotificationClass>;
  }

  function formatNotificationClassLabel(item: NotificationClass): string {
    return `NC ${item.nc} - ${item.object_description}`;
  }

  const sortedBacnetObjects = $derived(
    [...bacnetObjects].sort((a, b) => {
      const softwareTypeCompare = a.software_type.localeCompare(b.software_type);
      if (softwareTypeCompare !== 0) return softwareTypeCompare;
      return a.software_number - b.software_number;
    })
  );
  const undoFieldTitle = $derived($t('field_device.editing.undo_field'));
  const undoBacnetRowTitle = $derived($t('field_device.editing.undo_bacnet_object'));
</script>

{#if bacnetObjects.length > 0}
  <div class="max-w-full min-w-0 overflow-x-auto">
    <table use:resizableTableColumns class="w-max min-w-full table-auto text-sm">
      <thead>
        <tr class="border-b text-left text-xs text-muted-foreground">
          <th class="w-fit max-w-80 min-w-max pr-2 pb-2">
            {$t('field_device.bacnet.table.text_fix')}
          </th>
          <th data-table-resize-min-width="52" class="w-13 min-w-13 pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.alarms')}
          </th>
          <th data-table-resize-min-width="96" class="w-24 min-w-24 pr-2 pb-2">
            {$t('field_device.bacnet.table.state_text')}
          </th>
          <th
            data-table-resize-min-width="120"
            class="w-fit max-w-48 min-w-30 overflow-hidden pr-2 pb-2"
            title={$t('field_device.bacnet.table.notification_class')}
          >
            <span class="block truncate">{$t('field_device.bacnet.table.notification_class')}</span>
          </th>
          <th class="w-fit max-w-96 min-w-max pr-2 pb-2">
            {$t('field_device.bacnet.table.description')}
          </th>
          <th data-table-resize-min-width="90" class="w-22.5 min-w-22.5 pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.software')}
          </th>
          <th data-table-resize-min-width="90" class="w-22.5 min-w-22.5 pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.hardware')}
          </th>
          <th data-table-resize-min-width="72" class="w-18 min-w-18 pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.gms_visible')}
          </th>
          <th data-table-resize-min-width="72" class="w-18 min-w-18 pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.optional')}
          </th>
          <th class="w-fit max-w-80 min-w-max pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.text_individual')}
          </th>
          <th data-table-resizable="false" class="w-12 min-w-12 pr-2 pb-2"></th>
        </tr>
      </thead>
      <tbody>
        {#each sortedBacnetObjects as obj (obj.id)}
          <tr class="group/bacnet-row border-b border-border/60 last:border-0">
            <td class="bacnet-content-width-cell w-fit max-w-80 py-1 pr-1">
              <div
                class={getCollaborationClass(obj.id, 'text_fix')}
                title={getPreviewTitle(obj.id, 'text_fix')}
                {...editCell(obj.id, 'text_fix')}
              >
                <EditableCell
                  value={obj.text_fix}
                  pendingValue={getPendingTextValue(obj.id, 'text_fix', obj.text_fix)}
                  maxlength={250}
                  isDirty={isDirty(obj.id, 'text_fix')}
                  error={getFieldError(obj.id, 'text_fix')}
                  {disabled}
                  onSave={(v) => onEdit(obj.id, 'text_fix', v)}
                  undoTitle={undoFieldTitle}
                  onUndo={() => onUndoField?.(obj.id, 'text_fix')}
                />
              </div>
            </td>
            <td
              class="w-13 min-w-13 py-1 pr-1 text-center align-top"
              {...focusCell(obj.id, 'alarms')}
            >
              <Popover.Root>
                <Popover.Trigger>
                  {#snippet child({ props })}
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-7 w-7"
                      disabled={!hasAlarmType(obj)}
                      title={hasAlarmType(obj)
                        ? $t('field_device.bacnet.table.show_alarms')
                        : $t('field_device.bacnet.table.no_alarms')}
                      {...props}
                    >
                      <BellRing class="h-4 w-4" />
                    </Button>
                  {/snippet}
                </Popover.Trigger>
                <Popover.Content
                  class="max-h-[70vh] w-[24rem] overflow-y-auto p-2"
                  align="start"
                  side="right"
                >
                  {#if hasAlarmType(obj)}
                    <BacnetAlarmValuesEditor bacnetObjectId={obj.id} />
                  {:else}
                    <p class="text-xs text-muted-foreground">
                      {$t('field_device.bacnet.table.no_alarms')}
                    </p>
                  {/if}
                </Popover.Content>
              </Popover.Root>
            </td>
            <td class="w-fit max-w-24 min-w-24 overflow-hidden py-1 pr-1 align-top">
              <div
                class={[
                  'group/undo relative',
                  isDirty(obj.id, 'state_text_id') ? 'rounded-md ring-1 ring-ring' : '',
                  getCollaborationClass(obj.id, 'state_text_id')
                ]
                  .filter(Boolean)
                  .join(' ')}
                title={getPreviewTitle(obj.id, 'state_text_id')}
                {...focusCell(obj.id, 'state_text_id')}
              >
                <AsyncCombobox
                  value={getPendingIdValue(obj.id, 'state_text_id', obj.state_text_id)}
                  fetcher={fetchStateTexts}
                  fetchById={fetchStateTextById}
                  labelKey="ref_number"
                  labelFormatter={formatStateTextLabel}
                  itemTitleFormatter={formatStateTextTooltip}
                  placeholder={$t('field_device.bacnet.row.select')}
                  searchPlaceholder={$t('common.search')}
                  width="w-fit max-w-24 min-w-[90px]"
                  {disabled}
                  onValueChange={(value) => onEdit(obj.id, 'state_text_id', value || undefined)}
                />
                {#if isDirty(obj.id, 'state_text_id')}
                  <InlineUndoButton
                    title={undoFieldTitle}
                    onclick={() => onUndoField?.(obj.id, 'state_text_id')}
                  />
                {/if}
              </div>
              {#if getFieldError(obj.id, 'state_text_id')}
                <p class="mt-1 text-xs text-destructive">
                  {getFieldError(obj.id, 'state_text_id')}
                </p>
              {/if}
            </td>
            <td class="w-fit max-w-48 min-w-30 overflow-hidden py-1 pr-1 align-top">
              <div
                class={[
                  'group/undo relative',
                  isDirty(obj.id, 'notification_class_id') ? 'rounded-md ring-1 ring-ring' : '',
                  getCollaborationClass(obj.id, 'notification_class_id')
                ]
                  .filter(Boolean)
                  .join(' ')}
                title={getPreviewTitle(obj.id, 'notification_class_id')}
                {...focusCell(obj.id, 'notification_class_id')}
              >
                <AsyncCombobox
                  value={getPendingIdValue(
                    obj.id,
                    'notification_class_id',
                    obj.notification_class_id
                  )}
                  fetcher={fetchNotificationClasses}
                  fetchById={fetchNotificationClassById}
                  labelKey="nc"
                  labelFormatter={formatNotificationClassLabel}
                  placeholder={$t('field_device.bacnet.row.select')}
                  searchPlaceholder={$t('common.search')}
                  width="w-fit max-w-48 min-w-28"
                  popupWidth="w-[360px]"
                  {disabled}
                  onValueChange={(value) =>
                    onEdit(obj.id, 'notification_class_id', value || undefined)}
                />
                {#if isDirty(obj.id, 'notification_class_id')}
                  <InlineUndoButton
                    title={undoFieldTitle}
                    onclick={() => onUndoField?.(obj.id, 'notification_class_id')}
                  />
                {/if}
              </div>
              {#if getFieldError(obj.id, 'notification_class_id')}
                <p class="mt-1 text-xs text-destructive">
                  {getFieldError(obj.id, 'notification_class_id')}
                </p>
              {/if}
            </td>
            <td class="bacnet-content-width-cell w-fit max-w-96 py-1 pr-1">
              <div
                class={getCollaborationClass(obj.id, 'description')}
                title={getPreviewTitle(obj.id, 'description')}
                {...editCell(obj.id, 'description')}
              >
                <EditableCell
                  value={obj.description || ''}
                  pendingValue={getPendingTextValue(obj.id, 'description', obj.description || '')}
                  maxlength={250}
                  isDirty={isDirty(obj.id, 'description')}
                  error={getFieldError(obj.id, 'description')}
                  {disabled}
                  onSave={(v) => onEdit(obj.id, 'description', v || undefined)}
                  undoTitle={undoFieldTitle}
                  onUndo={() => onUndoField?.(obj.id, 'description')}
                />
              </div>
            </td>
            <td class="w-22.5 min-w-22.5 py-1 pr-1 align-top">
              <div
                class={`bacnet-inline-pair mx-auto inline-flex items-center ${getCollaborationClass(obj.id, 'software_type')} ${getCollaborationClass(obj.id, 'software_number')}`}
                title={[
                  getPreviewTitle(obj.id, 'software_type'),
                  getPreviewTitle(obj.id, 'software_number')
                ]
                  .filter(Boolean)
                  .join('\n') || undefined}
              >
                <div class="bacnet-inline-type" {...focusCell(obj.id, 'software_type')}>
                  <EditableSelectCell
                    value={obj.software_type}
                    options={softwareTypeOptions}
                    pendingValue={getPendingTextValue(obj.id, 'software_type', obj.software_type)}
                    isDirty={isDirty(obj.id, 'software_type')}
                    error={getFieldError(obj.id, 'software_type')}
                    {disabled}
                    onSave={(v) => onEdit(obj.id, 'software_type', v)}
                    undoTitle={undoFieldTitle}
                    onUndo={() => onUndoField?.(obj.id, 'software_type')}
                  />
                </div>
                <div class="bacnet-inline-number" {...editCell(obj.id, 'software_number')}>
                  <EditableCell
                    value={String(obj.software_number).padStart(2, '0')}
                    pendingValue={getPendingTextValue(
                      obj.id,
                      'software_number',
                      String(obj.software_number).padStart(2, '0')
                    )}
                    type="number"
                    min={1}
                    max={99}
                    isDirty={isDirty(obj.id, 'software_number')}
                    error={getFieldError(obj.id, 'software_number')}
                    {disabled}
                    onSave={(v) => {
                      const n = v ? Math.max(1, Math.min(99, parseInt(v))) : 1;
                      onEdit(obj.id, 'software_number', n);
                    }}
                    undoTitle={undoFieldTitle}
                    onUndo={() => onUndoField?.(obj.id, 'software_number')}
                  />
                </div>
              </div>
            </td>
            <td class="w-22.5 min-w-22.5 py-1 pr-1 align-top">
              <div
                class={`bacnet-inline-pair mx-auto inline-flex items-center ${getCollaborationClass(obj.id, 'hardware_type')} ${getCollaborationClass(obj.id, 'hardware_quantity')}`}
                title={[
                  getPreviewTitle(obj.id, 'hardware_type'),
                  getPreviewTitle(obj.id, 'hardware_quantity')
                ]
                  .filter(Boolean)
                  .join('\n') || undefined}
              >
                <div class="bacnet-inline-type" {...focusCell(obj.id, 'hardware_type')}>
                  <EditableSelectCell
                    value={obj.hardware_type}
                    options={hardwareTypeOptions}
                    pendingValue={getPendingTextValue(obj.id, 'hardware_type', obj.hardware_type)}
                    isDirty={isDirty(obj.id, 'hardware_type')}
                    error={getFieldError(obj.id, 'hardware_type')}
                    {disabled}
                    onSave={(v) => onEdit(obj.id, 'hardware_type', v)}
                    undoTitle={undoFieldTitle}
                    onUndo={() => onUndoField?.(obj.id, 'hardware_type')}
                  />
                </div>
                <div class="bacnet-inline-number" {...editCell(obj.id, 'hardware_quantity')}>
                  <EditableCell
                    value={String(obj.hardware_quantity).padStart(2, '0')}
                    pendingValue={getPendingTextValue(
                      obj.id,
                      'hardware_quantity',
                      String(obj.hardware_quantity).padStart(2, '0')
                    )}
                    type="number"
                    min={1}
                    max={99}
                    isDirty={isDirty(obj.id, 'hardware_quantity')}
                    error={getFieldError(obj.id, 'hardware_quantity')}
                    {disabled}
                    onSave={(v) => {
                      const n = v ? Math.max(1, Math.min(99, parseInt(v))) : 1;
                      onEdit(obj.id, 'hardware_quantity', n);
                    }}
                    undoTitle={undoFieldTitle}
                    onUndo={() => onUndoField?.(obj.id, 'hardware_quantity')}
                  />
                </div>
              </div>
            </td>
            <td class="w-18 min-w-18 py-1 pr-1">
              <div
                class={getCollaborationClass(obj.id, 'gms_visible')}
                title={getPreviewTitle(obj.id, 'gms_visible')}
                {...focusCell(obj.id, 'gms_visible')}
              >
                <EditableBooleanCell
                  value={obj.gms_visible}
                  pendingValue={getPendingBoolValue(obj.id, 'gms_visible')}
                  isDirty={isDirty(obj.id, 'gms_visible')}
                  error={getFieldError(obj.id, 'gms_visible')}
                  {disabled}
                  onToggle={(v) => onEdit(obj.id, 'gms_visible', v)}
                  undoTitle={undoFieldTitle}
                  onUndo={() => onUndoField?.(obj.id, 'gms_visible')}
                />
              </div>
            </td>
            <td class="w-18 min-w-18 py-1 pr-1">
              <div
                class={getCollaborationClass(obj.id, 'optional')}
                title={getPreviewTitle(obj.id, 'optional')}
                {...focusCell(obj.id, 'optional')}
              >
                <EditableBooleanCell
                  value={obj.optional}
                  pendingValue={getPendingBoolValue(obj.id, 'optional')}
                  isDirty={isDirty(obj.id, 'optional')}
                  error={getFieldError(obj.id, 'optional')}
                  {disabled}
                  onToggle={(v) => onEdit(obj.id, 'optional', v)}
                  undoTitle={undoFieldTitle}
                  onUndo={() => onUndoField?.(obj.id, 'optional')}
                />
              </div>
            </td>
            <td class="bacnet-content-width-cell w-fit max-w-80 py-1">
              {#if hasTextIndividual(obj)}
                {@const pendingTextIndividual = getPendingTextValue(
                  obj.id,
                  'text_individual',
                  obj.text_individual || ''
                )}
                {@const hasExistingTextIndividual =
                  (pendingTextIndividual ?? obj.text_individual ?? '').trim().length > 0}
                <div
                  class={getCollaborationClass(obj.id, 'text_individual')}
                  title={getPreviewTitle(obj.id, 'text_individual')}
                  {...editCell(obj.id, 'text_individual')}
                >
                  <EditableCell
                    value={obj.text_individual || ''}
                    pendingValue={pendingTextIndividual}
                    maxlength={250}
                    isDirty={isDirty(obj.id, 'text_individual')}
                    error={getFieldError(obj.id, 'text_individual')}
                    {disabled}
                    onSave={(v) => {
                      const normalized = v.trim();
                      onEdit(
                        obj.id,
                        'text_individual',
                        normalized === ''
                          ? hasExistingTextIndividual
                            ? ''
                            : undefined
                          : normalized
                      );
                    }}
                    undoTitle={undoFieldTitle}
                    onUndo={() => onUndoField?.(obj.id, 'text_individual')}
                  />
                </div>
              {/if}
            </td>
            <td class="w-12 min-w-12 py-1 pr-1 text-right align-top">
              {#if hasObjectEdits(obj.id)}
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  class="h-7 w-7 opacity-0 transition-opacity group-hover/bacnet-row:opacity-100 focus-visible:opacity-100"
                  title={undoBacnetRowTitle}
                  onclick={() => onUndoRow?.(obj.id)}
                >
                  <Undo2 class="size-4" />
                </Button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{:else}
  <p class="text-sm text-muted-foreground italic">
    {$t('field_device.bacnet.empty')}
  </p>
{/if}

<style>
  .bacnet-inline-pair {
    font-variant-numeric: tabular-nums;
  }

  .bacnet-inline-type {
    width: 2.35rem;
    min-width: 2.35rem;
  }

  .bacnet-inline-number {
    width: 1.85rem;
    min-width: 1.85rem;
  }

  .bacnet-content-width-cell :global(button:not(.editable-cell-display)) {
    width: fit-content;
    max-width: 100%;
  }

  .bacnet-content-width-cell :global(.editable-cell-display),
  .bacnet-content-width-cell :global(.editable-cell-editor) {
    width: 100%;
    max-width: 100%;
  }

  .bacnet-content-width-cell :global(input) {
    min-width: 0;
  }

  .bacnet-content-width-cell :global(span),
  .bacnet-content-width-cell :global(code) {
    max-width: 100%;
  }

  .bacnet-inline-pair :global(button),
  .bacnet-inline-pair :global(select) {
    width: 100%;
    min-width: 0;
    height: 1.75rem;
    min-height: 1.75rem;
    justify-content: flex-start;
    border-radius: 0.125rem;
    padding-inline: 0.125rem;
    font-size: 0.875rem;
  }

  .bacnet-inline-pair :global(code) {
    max-width: none;
    border-radius: 0;
    background: transparent;
    padding: 0;
    font-family: inherit;
    font-size: inherit;
  }
</style>
