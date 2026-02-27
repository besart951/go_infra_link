<script lang="ts">
	import { onDestroy } from 'svelte';
	import { canPerform } from '$lib/utils/permissions.js';
	import ExcelUploadDropzone from '$lib/components/excel/ExcelUploadDropzone.svelte';
	import ExcelReadProgressCard from '$lib/components/excel/ExcelReadProgressCard.svelte';
	import ExcelSessionSummary from '$lib/components/excel/ExcelSessionSummary.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { StartExcelReadSessionUseCase } from '$lib/application/useCases/excel/startExcelReadSessionUseCase.js';
	import { ExcelWorkerReaderAdapter } from '$lib/infrastructure/excel/excelWorkerReaderAdapter.js';
	import type { ExcelReadSession } from '$lib/domain/excel/index.js';

	const readSessionUseCase = new StartExcelReadSessionUseCase(new ExcelWorkerReaderAdapter());

	let isReading = $state(false);
	let progressPercent = $state(0);
	let progressMessage = $state('Waiting for file...');
	let errorMessage = $state<string | null>(null);
	let preparedSession = $state<ExcelReadSession | null>(null);

	async function startReadSession(file: File): Promise<void> {
		isReading = true;
		errorMessage = null;
		preparedSession = null;
		progressPercent = 0;
		progressMessage = 'Preparing scanner...';

		try {
			const session = await readSessionUseCase.execute(file, (progress) => {
				progressPercent = progress.percent;
				progressMessage = progress.message;
			});

			preparedSession = session;
			progressPercent = 100;
			progressMessage = 'Scanner result ready.';
			addToast('Excel file loaded and ObjectData/BACnet result prepared.', 'success');
		} catch (error) {
			const message = error instanceof Error ? error.message : 'Failed to read Excel file.';
			if (message === 'Read session cancelled.') {
				progressMessage = 'Read cancelled.';
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
		readSessionUseCase.cancel();
		isReading = false;
		progressMessage = 'Read cancelled.';
		progressPercent = 0;
	}

	onDestroy(() => {
		readSessionUseCase.cancel();
	});
</script>

<svelte:head>
	<title>Excel Importer | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Excel Importer</h1>
		<p class="text-sm text-muted-foreground">
			Drop an Excel file to scan ObjectData and BACnet objects directly in the browser.
		</p>
	</div>

	{#if canPerform('create', 'objectdata')}
	<ExcelUploadDropzone disabled={isReading} onFileSelected={handleFileSelected} />

	<ExcelReadProgressCard
		{progressPercent}
		{progressMessage}
		{isReading}
		onCancel={cancelReadSession}
	/>
	{:else}
	<div class="rounded-lg border bg-muted p-4 text-center text-sm text-muted-foreground">
		You do not have permission to import Excel data.
	</div>
	{/if}

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
