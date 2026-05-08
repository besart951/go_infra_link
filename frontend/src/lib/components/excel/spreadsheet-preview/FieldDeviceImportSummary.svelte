<script lang="ts">
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportService } from './FieldDeviceImportService.svelte.js';
  import type { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';

  interface Props {
    service: FieldDeviceImportService;
    view: FieldDeviceImportViewState;
  }

  let { service, view }: Props = $props();
  const t = createTranslator();
</script>

{#if service.plan}
  <div class="mt-4 grid grid-cols-2 gap-2 text-xs md:grid-cols-5">
    <div class="rounded-md border p-2">
      <div class="text-muted-foreground">
        {$t('field_device.importer.summary.control_cabinet')}
      </div>
      <div class="truncate font-medium">
        {service.plan.controller.controlCabinetNr || $t('common.not_available')}
      </div>
    </div>
    <div class="rounded-md border p-2">
      <div class="text-muted-foreground">{$t('field_device.importer.summary.system_types')}</div>
      <div class="font-medium">{service.plan.controller.systemTypes.length}</div>
    </div>
    <div class="rounded-md border p-2">
      <div class="text-muted-foreground">{$t('field_device.importer.summary.field_devices')}</div>
      <div class="font-medium">{service.plan.fieldDeviceCount}</div>
    </div>
    <div class="rounded-md border p-2">
      <div class="text-muted-foreground">{$t('field_device.importer.summary.bacnet_objects')}</div>
      <div class="font-medium">{service.plan.bacnetObjectCount}</div>
    </div>
    <div class="rounded-md border p-2">
      <div class="text-muted-foreground">{$t('field_device.importer.summary.validation')}</div>
      <div class="font-medium">
        {$t('field_device.importer.summary.validation_counts', {
          errors: view.blockingDiagnostics.length,
          warnings: view.warningDiagnostics.length
        })}
      </div>
    </div>
  </div>
{/if}
