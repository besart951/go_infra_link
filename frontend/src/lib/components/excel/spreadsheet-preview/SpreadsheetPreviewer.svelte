<script lang="ts">
  import { onDestroy } from 'svelte';
  import Dropzone from './Dropzone.svelte';
  import WorksheetSelector from './WorksheetSelector.svelte';
  import DataGrid from './DataGrid.svelte';
  import FieldDeviceImportPanel from './FieldDeviceImportPanel.svelte';
  import { WorkbookService } from './WorkbookService.svelte.js';
  import { FieldDeviceImportService } from './FieldDeviceImportService.svelte.js';
  import { SheetJsWorkbookParser } from '$lib/infrastructure/excel/sheetJsWorkbookParser.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { VersionedFieldDeviceImportState } from './VersionedFieldDeviceImportState.svelte.js';
  import { isVersionedFieldDeviceArchive } from '$lib/infrastructure/api/versionedFieldDeviceImport.js';

  const t = createTranslator();
  const workbookService = new WorkbookService(new SheetJsWorkbookParser(), {
    visibleRowLimit: 500
  });
  const fieldDeviceImportService = new FieldDeviceImportService();
  const versionedImport = new VersionedFieldDeviceImportState();

  onDestroy(() => {
    fieldDeviceImportService.dispose();
  });

  async function handleFileSelected(file: File): Promise<void> {
    fieldDeviceImportService.clearTransform();
    if (isVersionedFieldDeviceArchive(file)) {
      workbookService.clear();
      versionedImport.selectArchive(file);
      addToast($t('excel.worksheet_preview.toasts.archive_selected'), 'success');
      return;
    }
    await workbookService.loadFile(file);

    if (workbookService.errorMessage) {
      versionedImport.clear();
      addToast(workbookService.errorMessage, 'error');
      return;
    }

    if (workbookService.workbook) {
      versionedImport.select(file, workbookService.sheetNames);
      addToast($t('excel.worksheet_preview.toasts.workbook_loaded'), 'success');
    }
  }

  function selectWorksheet(name: string): void {
    fieldDeviceImportService.clearTransform();
    workbookService.selectWorksheet(name);
  }

  async function importVersionedWorkbook(): Promise<void> {
    const result = await versionedImport.run();
    if (!result) {
      if (versionedImport.errorMessage) addToast(versionedImport.errorMessage, 'error');
      return;
    }
    const kind = (result.failed ?? 0) > 0 || (result.issues?.length ?? 0) > 0 ? 'error' : 'success';
    addToast(
      $t('excel.worksheet_preview.versioned.result', {
        imported: result.imported ?? 0,
        failed: result.failed ?? 0
      }),
      kind
    );
  }
</script>

<div class="flex flex-col gap-4">
  <Dropzone
    disabled={workbookService.isLoading}
    fileName={workbookService.workbook?.fileName ?? versionedImport.file?.name ?? null}
    onFileSelected={handleFileSelected}
  />

  {#if workbookService.errorMessage}
    <div
      class="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive"
    >
      {workbookService.errorMessage}
    </div>
  {/if}

  {#if workbookService.isLoading}
    <div class="rounded-lg border bg-muted/20 p-4 text-sm text-muted-foreground">
      {$t('excel.worksheet_preview.loading_workbook')}
    </div>
  {/if}

  {#if workbookService.workbook}
    <WorksheetSelector
      worksheets={workbookService.sheetNames}
      selectedWorksheetName={workbookService.selectedWorksheetName}
      disabled={workbookService.isLoading}
      onSelect={selectWorksheet}
    />
  {/if}

  {#if versionedImport.isVersioned}
    <section class="rounded-lg border bg-muted/20 p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 class="font-medium">{$t('excel.worksheet_preview.versioned.title')}</h3>
          <p class="text-sm text-muted-foreground">
            {$t('excel.worksheet_preview.versioned.description')}
          </p>
        </div>
        <Button
          type="button"
          disabled={versionedImport.isImporting}
          onclick={importVersionedWorkbook}
        >
          {versionedImport.isImporting
            ? $t('excel.worksheet_preview.versioned.importing')
            : $t('excel.worksheet_preview.versioned.import')}
        </Button>
      </div>
      {#if versionedImport.result}
        <p class="mt-3 text-sm">
          {$t('excel.worksheet_preview.versioned.result', {
            imported: versionedImport.result.imported ?? 0,
            failed: versionedImport.result.failed ?? 0
          })}
        </p>
        {#if versionedImport.result.issues?.length}
          <ul class="mt-2 max-h-48 list-disc overflow-auto pl-5 text-sm text-destructive">
            {#each versionedImport.result.issues as issue (`${issue.entity}-${issue.source_id}-${issue.field}-${issue.code}`)}
              <li>{issue.entity ?? 'import'} · {issue.field ?? issue.code}: {issue.message}</li>
            {/each}
          </ul>
        {/if}
      {/if}
    </section>
  {:else}
    <FieldDeviceImportPanel
      worksheet={workbookService.selectedWorksheet}
      service={fieldDeviceImportService}
    />
  {/if}

  {#if !versionedImport.isArchive}
    <DataGrid
      worksheet={workbookService.selectedWorksheet}
      rows={workbookService.displayRows}
      columnLabels={workbookService.columnLabels}
      isTruncated={workbookService.isPreviewTruncated}
      visibleRowLimit={workbookService.visibleRowLimit}
      cellMarkers={fieldDeviceImportService.cellMarkers}
    />
  {/if}
</div>
