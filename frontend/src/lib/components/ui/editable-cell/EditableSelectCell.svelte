<script lang="ts">
  /**
   * EditableSelectCell Component
   * Inline editable table cell with click-to-edit select dropdown
   */
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import InlineUndoButton from './InlineUndoButton.svelte';

  interface SelectOption {
    value: string;
    label: string;
  }

  interface Props {
    value: string;
    options: SelectOption[];
    pendingValue?: string;
    isDirty?: boolean;
    error?: string;
    disabled?: boolean;
    emptyText?: string;
    undoTitle?: string;
    onSave: (value: string) => void;
    onUndo?: () => void;
  }

  let {
    value,
    options,
    pendingValue,
    isDirty = false,
    error,
    disabled = false,
    emptyText = '-',
    undoTitle = 'Undo field change',
    onSave,
    onUndo
  }: Props = $props();

  let isEditing = $state(false);
  let selectElement: HTMLSelectElement | null = $state(null);

  const displayValue = $derived(pendingValue !== undefined ? pendingValue : value);
  const displayLabel = $derived(
    options.find((o) => o.value === displayValue)?.label || displayValue || emptyText
  );
  const hasError = $derived(!!error);
  const canUndo = $derived(isDirty && !!onUndo && !isEditing);

  function startEditing() {
    if (disabled) return;
    isEditing = true;
    setTimeout(() => selectElement?.focus(), 0);
  }

  function handleChange(e: Event) {
    const newValue = (e.target as HTMLSelectElement).value;
    isEditing = false;
    if (newValue !== displayValue) {
      onSave(newValue);
    }
  }

  function handleBlur() {
    isEditing = false;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      isEditing = false;
    }
  }
</script>

{#if isEditing}
  <div class="editable-cell-editor relative block w-full max-w-full min-w-0 align-middle">
    <Button
      type="button"
      tabindex={-1}
      aria-hidden="true"
      variant="ghost"
      pressEffect="none"
      class={[
        'editable-cell-display invisible h-7 min-h-7 w-full max-w-full min-w-0 shrink cursor-pointer items-center justify-start gap-0 rounded-sm px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
        hasError ? 'border' : ''
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <span class="truncate">{displayLabel}</span>
    </Button>
    <select
      bind:this={selectElement}
      data-keyboard-table-ignore
      value={displayValue}
      onchange={handleChange}
      onblur={handleBlur}
      onkeydown={handleKeydown}
      class={[
        'absolute inset-0 h-7 w-full min-w-0 rounded-sm border bg-background px-1.5 py-0.5 text-sm focus:ring-1 focus:ring-ring focus:outline-none',
        hasError ? 'border-destructive focus:ring-destructive' : 'border-input'
      ]
        .filter(Boolean)
        .join(' ')}
    >
      {#each options as opt (opt.value)}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  </div>
{:else if hasError}
  <div class="group/undo relative">
    <Tooltip.Provider>
      <Tooltip.Root>
        <Tooltip.Trigger>
          {#snippet child({ props })}
            <Button
              {...props}
              type="button"
              variant="ghost"
              pressEffect="none"
              onclick={startEditing}
              {disabled}
              class={[
                'editable-cell-display h-7 min-h-7 w-full shrink cursor-pointer items-center justify-start gap-0 rounded-sm border px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
                'border-destructive bg-destructive/10 hover:bg-destructive/20',
                disabled ? 'cursor-not-allowed opacity-50' : ''
              ]
                .filter(Boolean)
                .join(' ')}
            >
              <span class="truncate">{displayLabel}</span>
            </Button>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content side="top" class="max-w-xs bg-destructive text-destructive-foreground">
          <p>{error}</p>
        </Tooltip.Content>
      </Tooltip.Root>
    </Tooltip.Provider>
    {#if canUndo}
      <InlineUndoButton title={undoTitle} onclick={() => onUndo?.()} />
    {/if}
  </div>
{:else}
  <div class="group/undo relative">
    <Button
      type="button"
      variant="ghost"
      pressEffect="none"
      onclick={startEditing}
      {disabled}
      class={[
        'editable-cell-display h-7 min-h-7 w-full shrink cursor-pointer items-center justify-start gap-0 rounded-sm px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
        'hover:bg-muted/50 focus:bg-muted/50 focus:outline-none',
        isDirty ? 'bg-warning-muted dark:bg-warning-muted/60' : '',
        disabled ? 'cursor-not-allowed opacity-50' : ''
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <span class="truncate">{displayLabel}</span>
    </Button>
    {#if canUndo}
      <InlineUndoButton title={undoTitle} onclick={() => onUndo?.()} />
    {/if}
  </div>
{/if}
