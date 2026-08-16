<script lang="ts">
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Field from '$lib/components/ui/field/index.js';
  import type { DetailSaveStatus } from './DetailRealtimeStatus.svelte';

  let {
    label,
    value,
    type = 'text',
    disabled = false,
    description,
    placeholder = 'Nicht hinterlegt',
    minLength,
    maxLength,
    min,
    max,
    transform,
    onSave,
    onStatusChange
  }: {
    label: string;
    value?: string | number | null;
    type?: 'text' | 'number';
    disabled?: boolean;
    description?: string;
    placeholder?: string;
    minLength?: number;
    maxLength?: number;
    min?: number;
    max?: number;
    transform?: (value: string) => string;
    onSave?: (value: string | number) => Promise<void>;
    onStatusChange?: (status: DetailSaveStatus) => void;
  } = $props();

  let draft = $state('');
  let focused = $state(false);
  let saving = $state(false);
  let errorMessage = $state<string | null>(null);

  const normalizedValue = $derived(value === null || value === undefined ? '' : String(value));

  $effect(() => {
    if (!focused && !saving) {
      draft = normalizedValue;
      errorMessage = null;
    }
  });

  async function save(): Promise<void> {
    if (disabled || !onSave || saving || draft === normalizedValue) {
      if (!focused) onStatusChange?.('saved');
      return;
    }

    saving = true;
    errorMessage = null;
    onStatusChange?.('saving');
    try {
      const next = type === 'number' ? Number(draft) : draft;
      if (type === 'number' && !Number.isFinite(next)) {
        throw new Error('Bitte eine gültige Zahl eingeben.');
      }
      await onSave(next);
      onStatusChange?.('saved');
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : 'Speichern fehlgeschlagen.';
      onStatusChange?.('conflict');
    } finally {
      saving = false;
    }
  }

  function startEditing(): void {
    focused = true;
    if (!disabled) onStatusChange?.('editing');
  }

  async function finishEditing(): Promise<void> {
    focused = false;
    await save();
  }

  function handleInput(event: Event): void {
    const input = event.currentTarget as HTMLInputElement;
    draft = transform ? transform(input.value) : input.value;
  }
</script>

<Field.Field class="gap-2">
  <Field.Label>{label}</Field.Label>
  {#if disabled || !onSave}
    <p class="min-h-9 rounded-md border border-transparent px-3 py-2 text-sm text-foreground">
      {normalizedValue || placeholder}
    </p>
  {:else}
    <Field.Content>
      <Input
        aria-label={label}
        {type}
        bind:value={draft}
        {placeholder}
        minlength={minLength}
        maxlength={maxLength}
        {min}
        {max}
        disabled={saving}
        aria-invalid={errorMessage !== null}
        onfocus={startEditing}
        onblur={finishEditing}
        oninput={handleInput}
        onkeydown={(event) => {
          if (event.key === 'Enter') {
            event.currentTarget.blur();
          }
        }}
      />
      {#if errorMessage}
        <Field.Error>{errorMessage}</Field.Error>
      {/if}
    </Field.Content>
  {/if}
  {#if description}
    <Field.Description>{description}</Field.Description>
  {/if}
</Field.Field>
