<script lang="ts">
  /**
   * BacnetObjectsEditor Component
   * Editable BACnet objects table for inline editing within expanded field device rows
   */
  import {
    EditableCell,
    EditableSelectCell,
    EditableBooleanCell
  } from '$lib/components/ui/editable-cell/index.js';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import BacnetAlarmValuesEditor from './BacnetAlarmValuesEditor.svelte';
  import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
  import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
  import {
    BACNET_SOFTWARE_TYPES,
    BACNET_HARDWARE_TYPES
  } from '$lib/domain/facility/bacnet-object.js';
  import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
  import type { BacnetObjectInput } from '$lib/domain/facility/field-device.js';
  import type { StateText, NotificationClass } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { BellRing } from '@lucide/svelte';

  interface Props {
    bacnetObjects: BacnetObject[];
    pendingEdits: Map<string, Partial<BacnetObjectInput>>;
    fieldErrors: Map<string, Record<string, string>>;
    clientErrors: Map<string, Record<string, string>>;
    disabled?: boolean;
    onEdit: (objectId: string, field: string, value: unknown) => void;
  }

  let {
    bacnetObjects,
    pendingEdits,
    fieldErrors,
    clientErrors,
    disabled = false,
    onEdit
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

  function isDirty(objectId: string, field: string): boolean {
    const edits = pendingEdits.get(objectId);
    return edits ? field in edits : false;
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
    return stateTextRepository.get(id);
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
    return notificationClassRepository.get(id);
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
</script>

{#if bacnetObjects.length > 0}
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b text-left text-xs text-muted-foreground">
          <th class="pr-2 pb-2">{$t('field_device.bacnet.table.text_fix')}</th>
          <th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.alarms')}</th>
          <th class="pr-2 pb-2">{$t('field_device.bacnet.table.state_text')}</th>
          <th class="pr-2 pb-2">{$t('field_device.bacnet.table.notification_class')}</th>
          <th class="pr-2 pb-2">{$t('field_device.bacnet.table.description')}</th>
          <th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.software')}</th>
          <th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.hardware')}</th>
          <th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.gms_visible')}</th>
          <th class="pr-2 pb-2 text-center">{$t('field_device.bacnet.table.optional')}</th>
          <th class="pr-2 pb-2 text-center">
            {$t('field_device.bacnet.table.text_individual')}
          </th>
        </tr>
      </thead>
      <tbody>
        {#each sortedBacnetObjects as obj (obj.id)}
          <tr class="border-b border-border/60 last:border-0">
            <td class="py-1 pr-1">
              <EditableCell
                value={obj.text_fix}
                pendingValue={getPendingTextValue(obj.id, 'text_fix', obj.text_fix)}
                maxlength={250}
                isDirty={isDirty(obj.id, 'text_fix')}
                error={getFieldError(obj.id, 'text_fix')}
                {disabled}
                onSave={(v) => onEdit(obj.id, 'text_fix', v)}
              />
            </td>
            <td class="py-1 pr-1 text-center align-top">
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
            <td class="py-1 pr-1 align-top">
              <div class={isDirty(obj.id, 'state_text_id') ? 'rounded-md ring-1 ring-ring' : ''}>
                <AsyncCombobox
                  value={getPendingIdValue(obj.id, 'state_text_id', obj.state_text_id)}
                  fetcher={fetchStateTexts}
                  fetchById={fetchStateTextById}
                  labelKey="ref_number"
                  labelFormatter={formatStateTextLabel}
                  itemTitleFormatter={formatStateTextTooltip}
                  placeholder={$t('field_device.bacnet.row.select')}
                  searchPlaceholder={$t('common.search')}
                  width="w-[90px]"
                  {disabled}
                  onValueChange={(value) => onEdit(obj.id, 'state_text_id', value || undefined)}
                />
              </div>
              {#if getFieldError(obj.id, 'state_text_id')}
                <p class="mt-1 text-xs text-red-500">{getFieldError(obj.id, 'state_text_id')}</p>
              {/if}
            </td>
            <td class="py-1 pr-1 align-top">
              <div
                class={isDirty(obj.id, 'notification_class_id')
                  ? 'rounded-md ring-1 ring-ring'
                  : ''}
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
                  width="w-[220px]"
                  {disabled}
                  onValueChange={(value) =>
                    onEdit(obj.id, 'notification_class_id', value || undefined)}
                />
              </div>
              {#if getFieldError(obj.id, 'notification_class_id')}
                <p class="mt-1 text-xs text-red-500">
                  {getFieldError(obj.id, 'notification_class_id')}
                </p>
              {/if}
            </td>
            <td class="max-w-sm py-1 pr-1">
              <EditableCell
                value={obj.description || ''}
                pendingValue={getPendingTextValue(obj.id, 'description', obj.description || '')}
                maxlength={250}
                isDirty={isDirty(obj.id, 'description')}
                error={getFieldError(obj.id, 'description')}
                {disabled}
                onSave={(v) => onEdit(obj.id, 'description', v || undefined)}
              />
            </td>
            <td class="py-1 pr-1">
              <div class="flex">
                <EditableSelectCell
                  value={obj.software_type}
                  options={softwareTypeOptions}
                  pendingValue={getPendingTextValue(obj.id, 'software_type', obj.software_type)}
                  isDirty={isDirty(obj.id, 'software_type')}
                  error={getFieldError(obj.id, 'software_type')}
                  {disabled}
                  onSave={(v) => onEdit(obj.id, 'software_type', v)}
                />
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
                />
              </div>
            </td>
            <td class="py-1 pr-1">
              <div class="flex">
                <EditableSelectCell
                  value={obj.hardware_type}
                  options={hardwareTypeOptions}
                  pendingValue={getPendingTextValue(obj.id, 'hardware_type', obj.hardware_type)}
                  isDirty={isDirty(obj.id, 'hardware_type')}
                  error={getFieldError(obj.id, 'hardware_type')}
                  {disabled}
                  onSave={(v) => onEdit(obj.id, 'hardware_type', v)}
                />
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
                />
              </div>
            </td>
            <td class="py-1 pr-1">
              <EditableBooleanCell
                value={obj.gms_visible}
                pendingValue={getPendingBoolValue(obj.id, 'gms_visible')}
                isDirty={isDirty(obj.id, 'gms_visible')}
                error={getFieldError(obj.id, 'gms_visible')}
                {disabled}
                onToggle={(v) => onEdit(obj.id, 'gms_visible', v)}
              />
            </td>
            <td class="py-1 pr-1">
              <EditableBooleanCell
                value={obj.optional}
                pendingValue={getPendingBoolValue(obj.id, 'optional')}
                isDirty={isDirty(obj.id, 'optional')}
                error={getFieldError(obj.id, 'optional')}
                {disabled}
                onToggle={(v) => onEdit(obj.id, 'optional', v)}
              />
            </td>
            <td class="py-1">
              {#if hasTextIndividual(obj)}
                {@const pendingTextIndividual = getPendingTextValue(
                  obj.id,
                  'text_individual',
                  obj.text_individual || ''
                )}
                {@const hasExistingTextIndividual =
                  (pendingTextIndividual ?? obj.text_individual ?? '').trim().length > 0}
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
                      normalized === '' ? (hasExistingTextIndividual ? '' : undefined) : normalized
                    );
                  }}
                />
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
