<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import type { WorksheetPreview } from '$lib/domain/excel/index.js';
  import { GitBranch, LoaderCircle, Play, WandSparkles } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceImportService } from './FieldDeviceImportService.svelte.js';
  import { FieldDeviceImportViewState } from './fieldDeviceImportPresentation.js';
  import FieldDeviceImportDiagnostics from './FieldDeviceImportDiagnostics.svelte';
  import FieldDeviceImportReportAlert from './FieldDeviceImportReportAlert.svelte';
  import FieldDeviceImportSummary from './FieldDeviceImportSummary.svelte';
  import FieldDeviceImportTree from './FieldDeviceImportTree.svelte';

  interface Props {
    worksheet: WorksheetPreview | null;
    service: FieldDeviceImportService;
  }

  let { worksheet, service }: Props = $props();
  const t = createTranslator();
  const view = $derived(new FieldDeviceImportViewState(service, (key, params) => $t(key, params)));
</script>

<Tooltip.Provider>
  <div class="rounded-lg border bg-background p-4">
    <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
      <div class="min-w-0">
        <h3 class="flex items-center gap-2 text-sm font-semibold">
          <GitBranch class="size-4" />
          {$t('field_device.importer.title')}
        </h3>
        <p class="text-xs text-muted-foreground">
          {worksheet ? worksheet.name : $t('field_device.importer.no_worksheet')}
        </p>
      </div>

      <div class="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="outline"
          disabled={!worksheet || service.isTransforming || service.isImporting}
          onclick={() => service.transform(worksheet)}
        >
          {#if service.isTransforming}
            <LoaderCircle class="mr-2 size-4 animate-spin" />
          {:else}
            <WandSparkles class="mr-2 size-4" />
          {/if}
          {$t('field_device.importer.actions.transform')}
        </Button>

        <Button
          type="button"
          disabled={!service.canImport || service.isTransforming}
          onclick={() => service.importPlan()}
        >
          {#if service.isImporting}
            <LoaderCircle class="mr-2 size-4 animate-spin" />
          {:else}
            <Play class="mr-2 size-4" />
          {/if}
          {service.importReport?.status === 'partial' || service.importReport?.status === 'failed'
            ? $t('field_device.importer.actions.retry')
            : $t('field_device.importer.actions.import')}
        </Button>
      </div>
    </div>

    {#if service.transformError}
      <div
        class="mt-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive"
      >
        {service.transformError}
      </div>
    {/if}

    {#if service.plan}
      <FieldDeviceImportSummary {service} {view} />
      <FieldDeviceImportReportAlert report={service.importReport} {view} />
      <FieldDeviceImportDiagnostics {view} />
      <FieldDeviceImportTree {service} {view} />
    {:else}
      <div class="mt-4 rounded-md border border-dashed p-4 text-sm text-muted-foreground">
        {$t('field_device.importer.empty_prompt')}
      </div>
    {/if}
  </div>
</Tooltip.Provider>
