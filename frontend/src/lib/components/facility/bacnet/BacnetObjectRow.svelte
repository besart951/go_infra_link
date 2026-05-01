<script lang="ts">
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Trash2 } from '@lucide/svelte';
  import { BACNET_SOFTWARE_TYPES, BACNET_HARDWARE_TYPES } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import AlarmTypeSelect from '$lib/components/facility/selects/AlarmTypeSelect.svelte';
  import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
  import type { AlarmTypeField, BacnetObjectInput } from '$lib/domain/facility/index.js';

  type BacnetRowErrors = Partial<
    Record<
      | 'text_fix'
      | 'description'
      | 'software_type'
      | 'software_number'
      | 'hardware_type'
      | 'hardware_quantity',
      string
    >
  >;

  interface Props {
    index: number;
    obj: BacnetObjectInput;
    readOnly?: boolean;
    errors?: BacnetRowErrors;
    onRemove?: () => void;
    onUpdate?: (field: string, value: any) => void;
  }

  let {
    index,
    obj = $bindable(),
    readOnly = false,
    errors = {},
    onRemove = () => {},
    onUpdate = () => {}
  }: Props = $props();

  const t = createTranslator();

  let textIndividualEnabled = $state(!!obj.text_individual);
  let prevGmsVisible = $state<boolean | null>(null);
  let prevOptional = $state<boolean | null>(null);
  let prevAlarmTypeId = $state<string | null>(null);
  let alarmTypeFields = $state<AlarmTypeField[]>([]);
  let alarmTypeFieldsLoading = $state(false);
  let alarmTypeFieldsError = $state('');
  const requiredAlarmTypeFields = $derived(alarmTypeFields.filter((field) => field.is_required));

  $effect(() => {
    if (readOnly) return;
    if (prevGmsVisible === null) {
      prevGmsVisible = obj.gms_visible;
      return;
    }
    if (obj.gms_visible !== prevGmsVisible) {
      prevGmsVisible = obj.gms_visible;
      onUpdate('gms_visible', obj.gms_visible);
    }
  });

  $effect(() => {
    if (readOnly) return;
    if (prevOptional === null) {
      prevOptional = obj.optional;
      return;
    }
    if (obj.optional !== prevOptional) {
      prevOptional = obj.optional;
      onUpdate('optional', obj.optional);
    }
  });

  $effect(() => {
    if (readOnly) return;
    if (prevAlarmTypeId === null) {
      prevAlarmTypeId = obj.alarm_type_id ?? '';
      return;
    }
    const current = obj.alarm_type_id ?? '';
    if (current !== prevAlarmTypeId) {
      prevAlarmTypeId = current;
      onUpdate('alarm_type_id', current || null);
    }
  });

  $effect(() => {
    const selectedAlarmTypeId = obj.alarm_type_id?.trim() ?? '';
    if (!selectedAlarmTypeId) {
      alarmTypeFields = [];
      alarmTypeFieldsError = '';
      alarmTypeFieldsLoading = false;
      return;
    }

    alarmTypeFieldsLoading = true;
    alarmTypeFieldsError = '';

    void alarmTypeRepository
      .getWithFields(selectedAlarmTypeId)
      .then((alarmType) => {
        alarmTypeFields = [...(alarmType.fields ?? [])].sort(
          (a, b) => (a.display_order ?? 0) - (b.display_order ?? 0)
        );
      })
      .catch(() => {
        alarmTypeFields = [];
        alarmTypeFieldsError = $t('field_device.bacnet.row.alarm_type_fields_load_failed');
      })
      .finally(() => {
        alarmTypeFieldsLoading = false;
      });
  });

  $effect(() => {
    if (readOnly) return;
    const value = textIndividualEnabled ? $t('field_device.bacnet.row.text_individual_value') : '';
    if (obj.text_individual !== value) {
      obj.text_individual = value;
      onUpdate('text_individual', value);
    }
  });

  $effect(() => {
    if (obj.text_individual && !textIndividualEnabled) {
      textIndividualEnabled = true;
    }
    if (!obj.text_individual && textIndividualEnabled) {
      textIndividualEnabled = false;
    }
  });
</script>

<div class="grid grid-cols-12 gap-2 rounded-md border p-3">
  <!-- Row number and remove button -->
  <div class="col-span-12 mb-2 flex items-center justify-between">
    <h4 class="text-sm font-semibold text-muted-foreground">
      {$t('field_device.bacnet.row.title', { index: index + 1 })}
    </h4>
    {#if !readOnly}
      <Button variant="ghost" size="sm" onclick={onRemove} class="h-7 w-7 p-0">
        <Trash2 class="size-4 text-destructive" />
      </Button>
    {/if}
  </div>

  <!-- Text Fix -->
  <div class="col-span-12 space-y-1 md:col-span-6">
    <Label for="text_fix_{index}" class="text-xs">{$t('field_device.bacnet.row.text_fix')}</Label>
    <Input
      id="text_fix_{index}"
      bind:value={obj.text_fix}
      onchange={() => onUpdate('text_fix', obj.text_fix)}
      required
      maxlength={250}
      placeholder={$t('field_device.bacnet.row.text_fix_placeholder')}
      class="h-8 text-sm"
      disabled={readOnly}
    />
    {#if errors.text_fix}
      <p class="text-xs text-destructive">{errors.text_fix}</p>
    {/if}
  </div>

  <!-- Description -->
  <div class="col-span-12 space-y-1 md:col-span-6">
    <Label for="description_{index}" class="text-xs">
      {$t('field_device.bacnet.row.description')}
    </Label>
    <Input
      id="description_{index}"
      bind:value={obj.description}
      onchange={() => onUpdate('description', obj.description)}
      maxlength={250}
      placeholder={$t('field_device.bacnet.row.description_placeholder')}
      class="h-8 text-sm"
      disabled={readOnly}
    />
  </div>

  <!-- Software Group: Type + Number -->
  <div class="col-span-12 space-y-1 md:col-span-6">
    <Label class="text-xs">{$t('field_device.bacnet.row.software')}</Label>
    <div class="grid grid-cols-2 gap-2">
      <div class="space-y-1">
        <Label for="software_type_{index}" class="text-xs text-muted-foreground">
          {$t('field_device.bacnet.row.type')}
        </Label>
        <select
          id="software_type_{index}"
          bind:value={obj.software_type}
          onchange={() => onUpdate('software_type', obj.software_type)}
          required
          disabled={readOnly}
          class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
        >
          <option value="">{$t('field_device.bacnet.row.select')}</option>
          {#each BACNET_SOFTWARE_TYPES as type}
            <option value={type.value}>{type.label}</option>
          {/each}
        </select>
        {#if errors.software_type}
          <p class="text-xs text-destructive">{errors.software_type}</p>
        {/if}
      </div>
      <div class="space-y-1">
        <Label for="software_number_{index}" class="text-xs text-muted-foreground">
          {$t('field_device.bacnet.row.number')}
        </Label>
        <Input
          id="software_number_{index}"
          type="number"
          bind:value={obj.software_number}
          onchange={() => onUpdate('software_number', obj.software_number)}
          required
          min={0}
          max={65535}
          placeholder={$t('field_device.bacnet.row.software_number_placeholder')}
          class="h-8 text-sm"
          disabled={readOnly}
        />
        {#if errors.software_number}
          <p class="text-xs text-destructive">{errors.software_number}</p>
        {/if}
      </div>
    </div>
  </div>

  <!-- Hardware Group: Type + Quantity -->
  <div class="col-span-12 space-y-1 md:col-span-6">
    <Label class="text-xs">{$t('field_device.bacnet.row.hardware')}</Label>
    <div class="grid grid-cols-2 gap-2">
      <div class="space-y-1">
        <Label for="hardware_type_{index}" class="text-xs text-muted-foreground">
          {$t('field_device.bacnet.row.type')}
        </Label>
        <select
          id="hardware_type_{index}"
          bind:value={obj.hardware_type}
          onchange={() => onUpdate('hardware_type', obj.hardware_type)}
          disabled={readOnly}
          class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
        >
          <option value="">{$t('field_device.bacnet.row.select')}</option>
          {#each BACNET_HARDWARE_TYPES as type}
            <option value={type.value}>{type.label}</option>
          {/each}
        </select>
        {#if errors.hardware_type}
          <p class="text-xs text-destructive">{errors.hardware_type}</p>
        {/if}
      </div>
      <div class="space-y-1">
        <Label for="hardware_quantity_{index}" class="text-xs text-muted-foreground">
          {$t('field_device.bacnet.row.quantity')}
        </Label>
        <Input
          id="hardware_quantity_{index}"
          type="number"
          bind:value={obj.hardware_quantity}
          onchange={() => onUpdate('hardware_quantity', obj.hardware_quantity)}
          min={0}
          max={255}
          placeholder={$t('field_device.bacnet.row.hardware_quantity_placeholder')}
          class="h-8 text-sm"
          disabled={readOnly}
        />
        {#if errors.hardware_quantity}
          <p class="text-xs text-destructive">{errors.hardware_quantity}</p>
        {/if}
      </div>
    </div>
  </div>

  <!-- Checkboxes -->
  <div class="col-span-12 flex flex-wrap items-center gap-4 md:col-span-6">
    <div class="flex items-center gap-2">
      <Checkbox id="gms_visible_{index}" bind:checked={obj.gms_visible} disabled={readOnly} />
      <Label for="gms_visible_{index}" class="cursor-pointer text-xs">
        {$t('field_device.bacnet.row.gms_visible')}
      </Label>
    </div>
    <div class="flex items-center gap-2">
      <Checkbox id="optional_{index}" bind:checked={obj.optional} disabled={readOnly} />
      <Label for="optional_{index}" class="cursor-pointer text-xs">
        {$t('field_device.bacnet.row.optional')}
      </Label>
    </div>
    <div class="flex items-center gap-2">
      <Checkbox
        id="text_individual_{index}"
        bind:checked={textIndividualEnabled}
        disabled={readOnly}
      />
      <Label for="text_individual_{index}" class="cursor-pointer text-xs">
        {$t('field_device.bacnet.row.text_individual')}
      </Label>
    </div>
  </div>

  <!-- Alarm Type Section -->
  <div class="col-span-12 space-y-1 border-t pt-2 md:col-span-12">
    <Label class="text-xs">{$t('field_device.bacnet.row.alarm_type')}</Label>
    <div class="space-y-2">
      {#if readOnly}
        <Input
          value={obj.alarm_type_id || ''}
          disabled
          placeholder={$t('field_device.bacnet.row.no_alarm_type')}
          class="h-8 text-sm"
        />
      {:else}
        <AlarmTypeSelect bind:value={obj.alarm_type_id} width="w-full" />
      {/if}
      {#if obj.alarm_type_id && !readOnly}
        <div class="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            onclick={() => {
              obj.alarm_type_id = '';
            }}
            class="h-7 px-2 text-xs"
            title={$t('field_device.bacnet.row.alarm_type_remove')}
          >
            {$t('field_device.bacnet.row.alarm_type_remove')}
          </Button>
        </div>
      {/if}
    </div>

    {#if alarmTypeFieldsLoading}
      <p class="text-xs text-muted-foreground">
        {$t('field_device.bacnet.row.alarm_type_fields_loading')}
      </p>
    {:else if alarmTypeFieldsError}
      <p class="text-xs text-destructive">{alarmTypeFieldsError}</p>
    {:else if requiredAlarmTypeFields.length > 0}
      <div class="rounded-md border bg-muted/30 p-2">
        <p class="mb-1 text-xs font-medium text-muted-foreground">
          {$t('field_device.bacnet.row.required_fields')}
        </p>
        <div class="space-y-1">
          {#each requiredAlarmTypeFields as field (field.id)}
            <div class="flex items-center justify-between gap-2 text-xs">
              <span class="truncate">
                {field.alarm_field?.label ?? field.alarm_field_id}
                ({field.alarm_field?.data_type ?? 'unknown'})
              </span>
              <span class="shrink-0 text-muted-foreground">{$t('common.required')}</span>
            </div>
          {/each}
        </div>
      </div>
    {:else if obj.alarm_type_id}
      <p class="text-xs text-muted-foreground">
        {$t('field_device.bacnet.row.no_required_fields')}
      </p>
    {/if}
  </div>
</div>
