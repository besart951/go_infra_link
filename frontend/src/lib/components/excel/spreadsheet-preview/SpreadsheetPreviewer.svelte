<script lang="ts">
  import Dropzone from './Dropzone.svelte';
  import WorksheetSelector from './WorksheetSelector.svelte';
  import DataGrid from './DataGrid.svelte';
  import { WorkbookService } from './WorkbookService.svelte.js';
  import { SheetJsWorkbookParser } from '$lib/infrastructure/excel/sheetJsWorkbookParser.js';
  import { addToast } from '$lib/components/toast.svelte';

  const workbookService = new WorkbookService(new SheetJsWorkbookParser(), {
    visibleRowLimit: 500
  });

  async function handleFileSelected(file: File): Promise<void> {
    await workbookService.loadFile(file);

    if (workbookService.errorMessage) {
      addToast(workbookService.errorMessage, 'error');
      return;
    }

    if (workbookService.workbook) {
      addToast('Excel-Arbeitsmappe geladen.', 'success');
    }
  }
</script>

<div class="flex flex-col gap-4">
  <Dropzone
    disabled={workbookService.isLoading}
    fileName={workbookService.workbook?.fileName ?? null}
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
      Arbeitsmappe wird gelesen...
    </div>
  {/if}

  {#if workbookService.workbook}
    <WorksheetSelector
      worksheets={workbookService.sheetNames}
      selectedWorksheetName={workbookService.selectedWorksheetName}
      disabled={workbookService.isLoading}
      onSelect={(name) => workbookService.selectWorksheet(name)}
    />
  {/if}

  <DataGrid
    worksheet={workbookService.selectedWorksheet}
    rows={workbookService.displayRows}
    columnLabels={workbookService.columnLabels}
    isTruncated={workbookService.isPreviewTruncated}
    visibleRowLimit={workbookService.visibleRowLimit}
  />
</div>
