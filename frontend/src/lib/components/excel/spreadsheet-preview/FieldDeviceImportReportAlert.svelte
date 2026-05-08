<script lang="ts">
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportReport } from './FieldDeviceImportService.svelte.js';
  import type {
    FieldDeviceImportViewState,
    ImportStatusVisualKind
  } from './fieldDeviceImportPresentation.js';
  import ImportStatusIcon from './ImportStatusIcon.svelte';

  interface Props {
    report: FieldDeviceImportReport | null;
    view: FieldDeviceImportViewState;
  }

  let { report, view }: Props = $props();
  const t = createTranslator();

  function reportVisualKind(report: FieldDeviceImportReport): ImportStatusVisualKind {
    if (report.status === 'success') return 'success';
    if (report.status === 'partial') return 'delta';
    return 'failed';
  }
</script>

{#if report}
  <div class={`mt-4 rounded-md border p-3 text-sm ${view.reportClass(report)}`}>
    <div class="flex items-center gap-2 font-medium">
      <ImportStatusIcon kind={reportVisualKind(report)} />
      {view.reportMessage(report)}
    </div>
    {#if report.createdControlCabinetLabel || report.createdSpsControllerLabel}
      <p class="mt-1 text-xs">
        {$t('field_device.importer.report.created_entities', {
          controlCabinet: report.createdControlCabinetLabel ?? $t('common.not_available'),
          spsController: report.createdSpsControllerLabel ?? $t('common.not_available')
        })}
      </p>
    {/if}
  </div>
{/if}
