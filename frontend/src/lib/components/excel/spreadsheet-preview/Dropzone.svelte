<script lang="ts">
  import { FileSpreadsheet, Upload } from '@lucide/svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    disabled?: boolean;
    fileName?: string | null;
    onFileSelected: (file: File) => void | Promise<void>;
  }

  let { disabled = false, fileName = null, onFileSelected }: Props = $props();
  const t = createTranslator();

  let fileInput: HTMLInputElement | null = null;
  let isDragActive = $state(false);

  function openFileDialog(): void {
    if (disabled) return;
    fileInput?.click();
  }

  function handleKeydown(event: KeyboardEvent): void {
    if (event.key !== 'Enter' && event.key !== ' ') return;
    event.preventDefault();
    openFileDialog();
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
  class={`rounded-lg border-2 border-dashed p-6 transition-colors ${
    isDragActive ? 'border-primary bg-muted/40' : 'border-border bg-background'
  }`}
  role="button"
  tabindex="0"
  aria-label={$t('excel.worksheet_preview.dropzone.aria_label')}
  ondrop={handleDrop}
  ondragover={handleDragOver}
  ondragenter={handleDragOver}
  ondragleave={handleDragLeave}
  onkeydown={handleKeydown}
>
  <div class="flex flex-col items-center justify-center gap-3 text-center">
    <FileSpreadsheet class="size-10 text-muted-foreground" />
    <div>
      <p class="text-sm font-medium">{$t('excel.worksheet_preview.dropzone.title')}</p>
      <p class="text-xs text-muted-foreground">
        {fileName ?? $t('excel.worksheet_preview.dropzone.supported_formats')}
      </p>
    </div>
    <Button type="button" variant="outline" onclick={openFileDialog} {disabled}>
      <Upload class="mr-2 size-4" />
      {$t('excel.worksheet_preview.dropzone.choose_file')}
    </Button>
    <input
      class="hidden"
      type="file"
      accept=".xlsx,.xlsm,.csv"
      bind:this={fileInput}
      onchange={handleInputChange}
      {disabled}
    />
  </div>
</div>
