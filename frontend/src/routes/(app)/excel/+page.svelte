<script lang="ts">
  import { onDestroy } from 'svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import ExcelUploadDropzone from '$lib/components/excel/ExcelUploadDropzone.svelte';
  import ExcelReadProgressCard from '$lib/components/excel/ExcelReadProgressCard.svelte';
  import ExcelSessionSummary from '$lib/components/excel/ExcelSessionSummary.svelte';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import type { StartExcelReadSessionUseCase } from '$lib/application/useCases/excel/startExcelReadSessionUseCase.js';
  import type { ExcelReadSession } from '$lib/domain/excel/index.js';
  import { FileSpreadsheet, Table2 } from '@lucide/svelte';

  type SpreadsheetPreviewerModule = typeof import('$lib/components/excel/spreadsheet-preview/SpreadsheetPreviewer.svelte');

  let readSessionUseCase: StartExcelReadSessionUseCase | null = null;
  let readSessionGeneration = 0;
  let spreadsheetPreviewerModule = $state<SpreadsheetPreviewerModule | null>(null);
  let spreadsheetPreviewerLoad = $state<Promise<SpreadsheetPreviewerModule> | null>(null);

  let activeImporterTab = $state('object-importer');
  let isReading = $state(false);
  let progressPercent = $state(0);
  let progressMessage = $state('Warten auf Datei...');
  let errorMessage = $state<string | null>(null);
  let preparedSession = $state<ExcelReadSession | null>(null);

  async function getReadSessionUseCase(): Promise<StartExcelReadSessionUseCase> {
    if (readSessionUseCase) {
      return readSessionUseCase;
    }

    const [{ StartExcelReadSessionUseCase }, { ExcelWorkerReaderAdapter }] = await Promise.all([
      import('$lib/application/useCases/excel/startExcelReadSessionUseCase.js'),
      import('$lib/infrastructure/excel/excelWorkerReaderAdapter.js')
    ]);
    readSessionUseCase = new StartExcelReadSessionUseCase(new ExcelWorkerReaderAdapter());
    return readSessionUseCase;
  }

  function loadSpreadsheetPreviewer(): Promise<SpreadsheetPreviewerModule> {
    if (spreadsheetPreviewerModule) {
      return Promise.resolve(spreadsheetPreviewerModule);
    }

    spreadsheetPreviewerLoad ??= import(
      '$lib/components/excel/spreadsheet-preview/SpreadsheetPreviewer.svelte'
    ).then((module) => {
      spreadsheetPreviewerModule = module;
      return module;
    });
    return spreadsheetPreviewerLoad;
  }

  async function startReadSession(file: File): Promise<void> {
    const generation = ++readSessionGeneration;
    isReading = true;
    errorMessage = null;
    preparedSession = null;
    progressPercent = 0;
    progressMessage = 'Scanner wird vorbereitet...';

    try {
      const useCase = await getReadSessionUseCase();
      if (generation !== readSessionGeneration) return;

      const session = await useCase.execute(file, (progress) => {
        if (generation !== readSessionGeneration) return;
        progressPercent = progress.percent;
        progressMessage = progress.message;
      });
      if (generation !== readSessionGeneration) return;

      preparedSession = session;
      progressPercent = 100;
      progressMessage = 'Scanner-Ergebnis bereit.';
      addToast('Excel-Datei geladen und Objekt-/BACnet-Daten vorbereitet.', 'success');
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'Excel-Datei konnte nicht gelesen werden.';
      if (message === 'Lesevorgang abgebrochen.') {
        progressMessage = 'Lesevorgang abgebrochen.';
        return;
      }

      errorMessage = message;
      addToast(errorMessage, 'error');
    } finally {
      isReading = false;
    }
  }

  async function handleFileSelected(file: File): Promise<void> {
    await startReadSession(file);
  }

  function cancelReadSession(): void {
    readSessionGeneration += 1;
    readSessionUseCase?.cancel();
    isReading = false;
    progressMessage = 'Lesevorgang abgebrochen.';
    progressPercent = 0;
  }

  onDestroy(() => {
    readSessionUseCase?.cancel();
  });
</script>

<svelte:head>
  <title>Excel-Import | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">Excel-Import</h1>
    <p class="text-sm text-muted-foreground">
      Laden Sie eine Excel-Datei hoch, um Objektdaten und BACnet-Objekte direkt im Browser zu
      prüfen.
    </p>
  </div>

  {#if canPerform('create', 'objectdata')}
    <Tabs.Root bind:value={activeImporterTab}>
      <Tabs.List class="w-full justify-start overflow-x-auto sm:w-fit">
        <Tabs.Trigger value="object-importer" class="gap-2">
          <FileSpreadsheet class="size-4" />
          Objektdaten-Importer
        </Tabs.Trigger>
        <Tabs.Trigger value="worksheet-preview" class="gap-2">
          <Table2 class="size-4" />
          Arbeitsblatt-Vorschau
        </Tabs.Trigger>
      </Tabs.List>

      <Tabs.Content value="object-importer" class="mt-4">
        <div class="flex flex-col gap-6">
          <ExcelUploadDropzone disabled={isReading} onFileSelected={handleFileSelected} />

          <ExcelReadProgressCard
            {progressPercent}
            {progressMessage}
            {isReading}
            onCancel={cancelReadSession}
          />

          {#if errorMessage}
            <div
              class="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive"
            >
              {errorMessage}
            </div>
          {/if}

          {#if preparedSession}
            <ExcelSessionSummary session={preparedSession} />
          {/if}
        </div>
      </Tabs.Content>

      <Tabs.Content value="worksheet-preview" class="mt-4">
        {#if activeImporterTab === 'worksheet-preview'}
          {#await loadSpreadsheetPreviewer()}
            <div class="rounded-lg border bg-muted/20 p-4">
              <div class="h-10 animate-pulse rounded bg-muted"></div>
            </div>
          {:then module}
            {@const Previewer = module.default}
            <Previewer />
          {:catch}
            <div
              class="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive"
            >
              Arbeitsblatt-Vorschau konnte nicht geladen werden.
            </div>
          {/await}
        {/if}
      </Tabs.Content>
    </Tabs.Root>
  {:else}
    <div class="rounded-lg border bg-muted p-4 text-center text-sm text-muted-foreground">
      Sie haben keine Berechtigung, Excel-Daten zu importieren.
    </div>
  {/if}
</div>
