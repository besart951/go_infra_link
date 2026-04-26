<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Trash2 } from '@lucide/svelte';
  import type { FieldDeviceMultiCreateState } from './multi-create/FieldDeviceMultiCreateState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    state: FieldDeviceMultiCreateState;
    index: number;
  }

  let { state, index }: Props = $props();

  const t = createTranslator();

  const row = $derived(state.rows[index]);
  const error = $derived(state.rowErrors.get(index) ?? null);
  const placeholder = $derived(state.getPlaceholderForRow(index));
  const disabled = $derived(state.submitting);
  const hasApparatNrError = $derived(error?.field === 'apparat_nr');
  const hasRowError = $derived(!!error && error.field !== 'apparat_nr');

  function handleRemove(): void {
    state.removeRow(index);
  }

  function handleBmkInput(event: Event): void {
    state.handleRowBmkChange(index, (event.target as HTMLInputElement).value);
  }

  function handleDescriptionInput(event: Event): void {
    state.handleRowDescriptionChange(index, (event.target as HTMLInputElement).value);
  }

  function handleTextFixInput(event: Event): void {
    state.handleRowTextFixChange(index, (event.target as HTMLInputElement).value);
  }

  function handleApparatNrInput(event: Event): void {
    state.handleRowApparatNrChange(index, (event.target as HTMLInputElement).value);
  }
</script>

<Table.Row class={hasRowError ? 'bg-destructive/5' : ''}>
  <Table.Cell class="align-center w-14 text-center text-sm">
    {index + 1}
  </Table.Cell>

  <Table.Cell class="min-w-36 align-top">
    <div class="space-y-1">
      <Input
        id={`bmk-${index}`}
        value={row.bmk}
        oninput={handleBmkInput}
        maxlength={10}
        {disabled}
      />
      {#if hasRowError && error}
        <p class="text-xs text-destructive">{error.message}</p>
      {/if}
    </div>
  </Table.Cell>

  <Table.Cell class="min-w-56 align-top">
    <div class="space-y-1">
      <Input
        id={`description-${index}`}
        value={row.description}
        oninput={handleDescriptionInput}
        maxlength={250}
        {disabled}
      />
    </div>
  </Table.Cell>

  <Table.Cell class="min-w-56 align-top">
    <div class="space-y-1">
      <Input
        id={`text-fix-${index}`}
        value={row.textFix ?? ''}
        oninput={handleTextFixInput}
        maxlength={250}
        {disabled}
      />
    </div>
  </Table.Cell>

  <Table.Cell class="w-36 align-top">
    <div class="space-y-1">
      <Input
        id={`apparat-nr-${index}`}
        type="number"
        value={row.apparatNr?.toString() ?? ''}
        oninput={handleApparatNrInput}
        {placeholder}
        min={1}
        max={99}
        {disabled}
        class={hasApparatNrError ? 'border-destructive' : ''}
      />
      {#if hasApparatNrError && error}
        <p class="text-xs text-destructive">{error.message}</p>
      {/if}
    </div>
  </Table.Cell>

  <Table.Cell class="w-16 text-center align-top">
    <Button variant="ghost" size="sm" onclick={handleRemove} {disabled}>
      <Trash2 class="size-4 text-destructive" />
    </Button>
  </Table.Cell>
</Table.Row>
