<script lang="ts">
	import { CircleCheck } from '@lucide/svelte';
	import type { ExcelReadSession } from '$lib/domain/excel/index.js';

	interface Props {
		session: ExcelReadSession;
		formatFileSize: (sizeInBytes: number) => string;
	}

	let { session, formatFileSize }: Props = $props();
</script>

<div class="rounded-lg border bg-background p-4">
	<div class="mb-3 flex items-center gap-2">
		<CircleCheck class="size-4 text-primary" />
		<h2 class="text-sm font-semibold">Read session ready</h2>
	</div>
	<div class="grid gap-2 text-sm md:grid-cols-3">
		<div><span class="text-muted-foreground">File:</span> {session.fileName}</div>
		<div><span class="text-muted-foreground">Size:</span> {formatFileSize(session.fileSize)}</div>
		<div><span class="text-muted-foreground">Sheets:</span> {session.totalSheets}</div>
	</div>

	<div class="mt-4 space-y-3">
		{#each session.sheets as sheet}
			<div class="rounded-md border p-3">
				<div class="flex items-center justify-between text-sm">
					<strong>{sheet.name}</strong>
					<span class="text-muted-foreground"
						>Rows: {sheet.rowCount} · Cols: {sheet.columnCount}</span
					>
				</div>
				{#if sheet.headers.length > 0}
					<p class="mt-2 text-xs text-muted-foreground">Headers: {sheet.headers.join(', ')}</p>
				{/if}
			</div>
		{/each}
	</div>
</div>
