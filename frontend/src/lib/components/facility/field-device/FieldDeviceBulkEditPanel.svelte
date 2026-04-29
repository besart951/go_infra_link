<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { ChevronDown, ChevronRight, CopyCheck, Eraser } from '@lucide/svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import TableApparatSelect from '../table-selects/TableApparatSelect.svelte';
  import TableSystemPartSelect from '../table-selects/TableSystemPartSelect.svelte';
  import type {
    BulkUpdateFieldDeviceItem,
    SpecificationInput
  } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const fieldDeviceState = useFieldDeviceState();

  let bulkEditValues = $state<Partial<BulkUpdateFieldDeviceItem>>({});
  let bulkSpecValues = $state<Partial<SpecificationInput>>({});
  let showBulkSpecFields = $state(false);

  const canEditBaseFields = $derived(fieldDeviceState.canUpdateFieldDevice());
  const canEditSpecificationFields = $derived(fieldDeviceState.canUpdateFieldDeviceSpecification());

  const hasBulkValues = $derived.by(
    () =>
      Object.values(bulkEditValues).some((value) => value !== undefined && value !== '') ||
      Object.values(bulkSpecValues).some((value) => value !== undefined && value !== '')
  );

  function applyBulkEdits() {
    if (fieldDeviceState.selectedIds.size === 0) return;

    let appliedCount = 0;

    for (const deviceId of fieldDeviceState.selectedIds) {
      if (canEditBaseFields) {
        for (const [field, value] of Object.entries(bulkEditValues)) {
          if (value === undefined || value === '') continue;
          fieldDeviceState.editing.queueEdit(
            deviceId,
            field as keyof BulkUpdateFieldDeviceItem,
            value
          );
          appliedCount++;
        }
      }

      if (canEditSpecificationFields) {
        for (const [field, value] of Object.entries(bulkSpecValues)) {
          if (value === undefined || value === '') continue;
          fieldDeviceState.editing.queueSpecEdit(
            deviceId,
            field as keyof SpecificationInput,
            value
          );
          appliedCount++;
        }
      }
    }

    if (appliedCount > 0) {
      addToast(
        translate('field_device.bulk_edit.toasts.applied', {
          count: fieldDeviceState.selectedIds.size
        }),
        'success'
      );
      return;
    }

    addToast(translate('field_device.bulk_edit.toasts.no_fields'), 'error');
  }

  function clearBulkEdit() {
    bulkEditValues = {};
    bulkSpecValues = {};
  }
</script>

<Card.Root class="border-primary/30 bg-primary/5">
  <Card.Header class="pb-3">
    <Card.Title class="text-base">
      {$t('field_device.bulk_edit.title', { count: fieldDeviceState.selectedCount })}
    </Card.Title>
    <Card.Description>
      {$t('field_device.bulk_edit.description')}
    </Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="mb-4">
      <fieldset disabled={!canEditBaseFields}>
        <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-6">
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.bmk')}</Label>
            <Input
              type="text"
              placeholder={$t('field_device.bulk_edit.bmk_placeholder')}
              maxlength={10}
              value={bulkEditValues.bmk ?? ''}
              oninput={(event: Event) => {
                const value = (event.target as HTMLInputElement).value;
                bulkEditValues = { ...bulkEditValues, bmk: value || undefined };
              }}
            />
          </div>
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.description')}</Label>
            <Input
              type="text"
              placeholder={$t('field_device.bulk_edit.description_placeholder')}
              maxlength={250}
              value={bulkEditValues.description ?? ''}
              oninput={(event: Event) => {
                const value = (event.target as HTMLInputElement).value;
                bulkEditValues = { ...bulkEditValues, description: value || undefined };
              }}
            />
          </div>
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.text_fix')}</Label>
            <Input
              type="text"
              placeholder={$t('field_device.bulk_edit.text_fix_placeholder')}
              maxlength={250}
              value={bulkEditValues.text_fix ?? ''}
              oninput={(event: Event) => {
                const value = (event.target as HTMLInputElement).value;
                bulkEditValues = { ...bulkEditValues, text_fix: value || undefined };
              }}
            />
          </div>
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.apparat_nr')}</Label>
            <Input
              type="number"
              placeholder={$t('field_device.bulk_edit.apparat_nr_placeholder')}
              min={1}
              max={99}
              value={bulkEditValues.apparat_nr?.toString() ?? ''}
              oninput={(event: Event) => {
                const value = (event.target as HTMLInputElement).value;
                bulkEditValues = {
                  ...bulkEditValues,
                  apparat_nr: value ? parseInt(value, 10) : undefined
                };
              }}
            />
          </div>
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.apparat')}</Label>
            <TableApparatSelect
              items={fieldDeviceState.allApparats}
              value={bulkEditValues.apparat_id ?? ''}
              width="w-full"
              onValueChange={(value) => {
                bulkEditValues = { ...bulkEditValues, apparat_id: value || undefined };
              }}
            />
          </div>
          <div class="flex flex-col gap-1">
            <Label class="text-xs">{$t('field_device.bulk_edit.system_part')}</Label>
            <TableSystemPartSelect
              items={fieldDeviceState.allSystemParts}
              value={bulkEditValues.system_part_id ?? ''}
              width="w-full"
              onValueChange={(value) => {
                bulkEditValues = { ...bulkEditValues, system_part_id: value || undefined };
              }}
            />
          </div>
        </div>
      </fieldset>
    </div>

    <div class="mb-4">
      {#if canEditSpecificationFields}
        <button
          type="button"
          class="mb-2 flex items-center gap-1 text-sm font-medium hover:underline"
          onclick={() => (showBulkSpecFields = !showBulkSpecFields)}
        >
          {#if showBulkSpecFields}
            <ChevronDown class="h-4 w-4" />
          {:else}
            <ChevronRight class="h-4 w-4" />
          {/if}
          {$t('field_device.bulk_edit.spec_fields')}
        </button>
      {/if}
      {#if showBulkSpecFields && canEditSpecificationFields}
        <fieldset>
          <div class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4">
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.supplier')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.supplier_placeholder')}
                maxlength={250}
                value={bulkSpecValues.specification_supplier ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    specification_supplier: value || undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.brand')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.brand_placeholder')}
                maxlength={250}
                value={bulkSpecValues.specification_brand ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = { ...bulkSpecValues, specification_brand: value || undefined };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.type')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.type_placeholder')}
                maxlength={250}
                value={bulkSpecValues.specification_type ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = { ...bulkSpecValues, specification_type: value || undefined };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.motor_valve')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.motor_valve_placeholder')}
                maxlength={250}
                value={bulkSpecValues.additional_info_motor_valve ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    additional_info_motor_valve: value || undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.size')}</Label>
              <Input
                type="number"
                placeholder={$t('field_device.bulk_edit.size_placeholder')}
                value={bulkSpecValues.additional_info_size?.toString() ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    additional_info_size: value ? parseInt(value, 10) : undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.install_location')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.install_location_placeholder')}
                maxlength={250}
                value={bulkSpecValues.additional_information_installation_location ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    additional_information_installation_location: value || undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.ph')}</Label>
              <Input
                type="number"
                placeholder={$t('field_device.bulk_edit.ph_placeholder')}
                value={bulkSpecValues.electrical_connection_ph?.toString() ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    electrical_connection_ph: value ? parseInt(value, 10) : undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.acdc')}</Label>
              <Input
                type="text"
                placeholder={$t('field_device.bulk_edit.acdc_placeholder')}
                maxlength={2}
                value={bulkSpecValues.electrical_connection_acdc ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    electrical_connection_acdc: value || undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.amperage')}</Label>
              <Input
                type="number"
                placeholder={$t('field_device.bulk_edit.amperage_placeholder')}
                value={bulkSpecValues.electrical_connection_amperage?.toString() ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    electrical_connection_amperage: value ? parseFloat(value) : undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.power')}</Label>
              <Input
                type="number"
                placeholder={$t('field_device.bulk_edit.power_placeholder')}
                value={bulkSpecValues.electrical_connection_power?.toString() ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    electrical_connection_power: value ? parseFloat(value) : undefined
                  };
                }}
              />
            </div>
            <div class="flex flex-col gap-1">
              <Label class="text-xs">{$t('field_device.bulk_edit.rotation')}</Label>
              <Input
                type="number"
                placeholder={$t('field_device.bulk_edit.rotation_placeholder')}
                value={bulkSpecValues.electrical_connection_rotation?.toString() ?? ''}
                oninput={(event: Event) => {
                  const value = (event.target as HTMLInputElement).value;
                  bulkSpecValues = {
                    ...bulkSpecValues,
                    electrical_connection_rotation: value ? parseInt(value, 10) : undefined
                  };
                }}
              />
            </div>
          </div>
        </fieldset>
      {/if}
    </div>

    <div class="flex gap-2">
      <Button size="sm" onclick={applyBulkEdits} disabled={!hasBulkValues}>
        <CopyCheck class="mr-1 h-4 w-4" />
        {$t('field_device.bulk_edit.apply')}
      </Button>
      <Button variant="outline" size="sm" onclick={clearBulkEdit} disabled={!hasBulkValues}>
        <Eraser class="mr-1 h-4 w-4" />
        {$t('field_device.bulk_edit.clear')}
      </Button>
    </div>
  </Card.Content>
</Card.Root>
