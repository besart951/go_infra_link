<script lang="ts">
	import { Upload, FileSpreadsheet } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		disabled?: boolean;
		onFileSelected: (file: File) => void | Promise<void>;
	}

	let { disabled = false, onFileSelected }: Props = $props();

	let fileInput: HTMLInputElement | null = null;
	let isDragActive = $state(false);

	function openFileDialog(): void {
		if (disabled) return;
		fileInput?.click();
	}

	function handleDropzoneKeydown(event: KeyboardEvent): void {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			openFileDialog();
		}
	}

	async function handleInputChange(event: Event): Promise<void> {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file || disabled) return;

		await onFileSelected(file);
		target.value = '';
	}

	async function handleDrop(event: DragEvent): Promise<void> {
		event.preventDefault();
		isDragActive = false;

		const file = event.dataTransfer?.files?.[0];
		if (!file || disabled) return;

		await onFileSelected(file);
	}

	function handleDragOver(event: DragEvent): void {
		event.preventDefault();
		if (disabled) return;
		isDragActive = true;
	}

	function handleDragLeave(event: DragEvent): void {
		event.preventDefault();
		isDragActive = false;
	}
</script>

<div
	class={`rounded-lg border-2 border-dashed p-8 transition-colors ${
		isDragActive ? 'border-primary bg-muted/40' : 'border-border bg-background'
	}`}
	role="button"
	tabindex="0"
	aria-label="Drop Excel file or open file picker"
	ondrop={handleDrop}
	ondragover={handleDragOver}
	ondragenter={handleDragOver}
	ondragleave={handleDragLeave}
	onkeydown={handleDropzoneKeydown}
>
	<div class="flex flex-col items-center justify-center gap-3 text-center">
		<FileSpreadsheet class="size-10 text-muted-foreground" />
		<p class="text-sm font-medium">Drag & drop your Excel file here</p>
		<p class="text-xs text-muted-foreground">Supported formats: .xlsx, .xls, .xlsm, .xlsb</p>
		<Button type="button" onclick={openFileDialog} {disabled}>
			<Upload class="mr-2 size-4" />
			Choose file
		</Button>
		<input
			class="hidden"
			type="file"
			accept=".xlsx,.xls,.xlsm,.xlsb"
			bind:this={fileInput}
			onchange={handleInputChange}
			{disabled}
		/>
	</div>
</div>
