<script lang="ts">
  import { onDestroy } from 'svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import ExcelUploadDropzone from '$lib/components/excel/ExcelUploadDropzone.svelte';
  import ExcelReadProgressCard from '$lib/components/excel/ExcelReadProgressCard.svelte';
  import ExcelSessionSummary from '$lib/components/excel/ExcelSessionSummary.svelte';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import type { StartExcelReadSessionUseCase } from '$lib/application/useCases/excel/startExcelReadSessionUseCase.js';
  import type { ExcelReadSession } from '$lib/domain/excel/index.js';
  import { FileSpreadsheet, Table2 } from '@lucide/svelte';

  type SpreadsheetPreviewerModule =
    typeof import('$lib/components/excel/spreadsheet-preview/SpreadsheetPreviewer.svelte');

  const t = createTranslator();

  let readSessionUseCase: StartExcelReadSessionUseCase | null = null;
  let readSessionGeneration = 0;
  let spreadsheetPreviewerModule = $state<SpreadsheetPreviewerModule | null>(null);
  let spreadsheetPreviewerLoad: Promise<void> | null = null;
  let isSpreadsheetPreviewerLoading = $state(false);
  let spreadsheetPreviewerLoadFailed = $state(false);

  let activeImporterTab = $state('object-importer');
  let isReading = $state(false);
  let progressPercent = $state(0);
  let progressMessage = $state(translate('excel.import.progress.waiting'));
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

  function loadSpreadsheetPreviewer(): void {
    if (spreadsheetPreviewerModule || spreadsheetPreviewerLoad) return;

    isSpreadsheetPreviewerLoading = true;
    spreadsheetPreviewerLoadFailed = false;
    spreadsheetPreviewerLoad = import(
      '$lib/components/excel/spreadsheet-preview/SpreadsheetPreviewer.svelte'
    )
      .then((module) => {
        spreadsheetPreviewerModule = module;
      })
      .catch((error) => {
        console.error('Failed to load spreadsheet previewer:', error);
        spreadsheetPreviewerLoadFailed = true;
      })
      .finally(() => {
        isSpreadsheetPreviewerLoading = false;
        spreadsheetPreviewerLoad = null;
      });
  }

  async function startReadSession(file: File): Promise<void> {
    const generation = ++readSessionGeneration;
    isReading = true;
    errorMessage = null;
    preparedSession = null;
    progressPercent = 0;
    progressMessage = translate('excel.import.progress.preparing');

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
      progressMessage = translate('excel.import.progress.ready');
      addToast(translate('excel.import.toasts.loaded_prepared'), 'success');
    } catch (error) {
      const message =
        error instanceof Error ? error.message : translate('excel.import.errors.read_failed');
      if (message === translate('excel.import.progress.cancelled')) {
        progressMessage = translate('excel.import.progress.cancelled');
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
    progressMessage = translate('excel.import.progress.cancelled');
    progressPercent = 0;
  }

  onDestroy(() => {
    readSessionUseCase?.cancel();
  });

  $effect(() => {
    if (activeImporterTab === 'worksheet-preview') {
      loadSpreadsheetPreviewer();
    }
  });
</script>

<svelte:head>
  <title>{$t('excel.import.title')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">{$t('excel.import.title')}</h1>
    <p class="text-sm text-muted-foreground">
      {$t('excel.import.description')}
    </p>
  </div>

  {#if canPerform('create', 'objectdata')}
    <Tabs.Root bind:value={activeImporterTab}>
      <Tabs.List class="w-full justify-start overflow-x-auto sm:w-fit">
        <Tabs.Trigger value="object-importer" class="gap-2">
          <FileSpreadsheet class="size-4" />
          {$t('excel.import.tabs.object_importer')}
        </Tabs.Trigger>
        <Tabs.Trigger value="worksheet-preview" class="gap-2">
          <Table2 class="size-4" />
          {$t('excel.import.tabs.worksheet_preview')}
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
          {#if isSpreadsheetPreviewerLoading}
            <div class="rounded-lg border bg-muted/20 p-4">
              <div class="h-10 animate-pulse rounded bg-muted"></div>
            </div>
          {:else if spreadsheetPreviewerLoadFailed}
            <div
              class="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive"
            >
              {$t('excel.import.errors.preview_load_failed')}
            </div>
          {:else if spreadsheetPreviewerModule}
            {@const Previewer = spreadsheetPreviewerModule.default}
            <Previewer />
          {/if}
        {/if}
      </Tabs.Content>
    </Tabs.Root>
  {:else}
    <div class="rounded-lg border bg-muted p-4 text-center text-sm text-muted-foreground">
      {$t('excel.import.errors.no_permission')}
    </div>
  {/if}
</div>
