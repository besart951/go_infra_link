<script lang="ts">
  /**
   * EditableCell Component
   * Inline editable table cell with click-to-edit behavior
   * Supports pending values display and error states
   */
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import InlineUndoButton from './InlineUndoButton.svelte';

  interface Props {
    value: string;
    pendingValue?: string; // Value to display when there's a pending edit
    type?: 'text' | 'number';
    placeholder?: string;
    maxlength?: number;
    min?: number;
    max?: number;
    isDirty?: boolean;
    error?: string; // Error message to display
    suggestion?: string;
    suggestionLabel?: string;
    suggestionActionLabel?: string;
    disabled?: boolean;
    emptyText?: string;
    undoTitle?: string;
    onSave: (value: string) => void;
    onUndo?: () => void;
    onApplySuggestion?: (value: string) => void;
  }

  let {
    value,
    pendingValue,
    type = 'text',
    placeholder = '',
    maxlength,
    min,
    max,
    isDirty = false,
    error,
    suggestion,
    suggestionLabel,
    suggestionActionLabel,
    disabled = false,
    emptyText = '-',
    undoTitle = 'Undo field change',
    onSave,
    onUndo,
    onApplySuggestion
  }: Props = $props();

  let isEditing = $state(false);
  let editValue = $state('');
  let inputElement: HTMLInputElement | null = $state(null);

  // Display value: use pending value if available, otherwise original value
  const displayValue = $derived(pendingValue !== undefined ? pendingValue : value);
  const displayTitle = $derived(displayValue ? displayValue : undefined);
  const displaySizerValue = $derived(displayValue || emptyText);
  const hasError = $derived(!!error);
  const hasSuggestion = $derived(!!suggestion);
  const canUndo = $derived(isDirty && !!onUndo && !isEditing);
  const suggestionTitle = $derived(suggestionLabel || suggestionActionLabel);

  function startEditing() {
    if (disabled) return;
    // Start with the display value (pending or original)
    editValue = displayValue;
    isEditing = true;
    // Focus after DOM update
    setTimeout(() => {
      inputElement?.focus();
      inputElement?.select();
    }, 0);
  }

  function handleSave() {
    isEditing = false;
    if (editValue !== displayValue) {
      onSave(editValue);
    }
  }

  function handleCancel() {
    isEditing = false;
    editValue = displayValue;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSave();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      handleCancel();
    }
  }

  function handleBlur() {
    handleSave();
  }

  // Update editValue when display value changes
  $effect(() => {
    if (!isEditing) {
      editValue = displayValue;
    }
  });
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
        'editable-cell-display invisible h-7 min-h-7 w-full max-w-full min-w-0 shrink cursor-pointer items-center justify-start gap-0 overflow-hidden rounded-sm px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
        hasError ? 'border' : ''
      ]
        .filter(Boolean)
        .join(' ')}
    >
      {#if type === 'number'}
        <code class="block max-w-full truncate rounded-md bg-muted px-1.5 py-0.5 text-sm">
          {displaySizerValue}
        </code>
      {:else}
        <span class="min-w-0 truncate">{displaySizerValue}</span>
      {/if}
    </Button>
    <Input
      bind:ref={inputElement}
      {type}
      bind:value={editValue}
      {placeholder}
      {maxlength}
      {min}
      {max}
      onkeydown={handleKeydown}
      onblur={handleBlur}
      class={[
        'absolute inset-0 h-7 w-full min-w-0 px-2 py-1 text-sm',
        hasError ? 'border-destructive focus-visible:ring-destructive' : ''
      ]
        .filter(Boolean)
        .join(' ')}
    />
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
                'editable-cell-display h-7 min-h-7 w-full min-w-0 shrink cursor-pointer items-center justify-start gap-0 overflow-hidden rounded-sm border px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
                'border-destructive bg-destructive/10 hover:bg-destructive/20',
                disabled ? 'cursor-not-allowed opacity-50' : ''
              ]
                .filter(Boolean)
                .join(' ')}
            >
              {#if displayValue}
                {#if type === 'number'}
                  <code class="block max-w-full truncate rounded-md bg-muted px-1.5 py-0.5 text-sm">
                    {displayValue}
                  </code>
                {:else}
                  <span class="min-w-0 truncate">{displayValue}</span>
                {/if}
              {:else}
                <span class="min-w-0 truncate text-muted-foreground">{emptyText}</span>
              {/if}
            </Button>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content side="top" class="max-w-xs bg-destructive text-destructive-foreground">
          <p>{error}</p>
          {#if hasSuggestion && suggestionLabel}
            <p class="mt-1 text-xs opacity-90">{suggestionLabel}</p>
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    </Tooltip.Provider>
    {#if hasSuggestion}
      {#if onApplySuggestion}
        <Button
          type="button"
          variant="ghost"
          pressEffect="none"
          class="mt-1 block max-w-full truncate rounded-sm bg-destructive/10 px-1 py-0.5 text-left text-[10px] font-medium text-destructive underline-offset-2 hover:underline focus:ring-1 focus:ring-destructive focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          title={suggestionTitle}
          {disabled}
          onclick={() => suggestion && onApplySuggestion(suggestion)}
        >
          {suggestionActionLabel ?? suggestionLabel ?? suggestion}
        </Button>
      {:else}
        <span
          class="mt-1 block max-w-full truncate rounded-sm bg-destructive/10 px-1 py-0.5 text-[10px] font-medium text-destructive"
          title={suggestionTitle}
        >
          {suggestionActionLabel ?? suggestionLabel ?? suggestion}
        </span>
      {/if}
    {/if}
    {#if canUndo}
      <InlineUndoButton title={undoTitle} onclick={() => onUndo?.()} />
    {/if}
  </div>
{:else if displayTitle}
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
                'editable-cell-display h-7 min-h-7 w-full min-w-0 shrink cursor-pointer items-center justify-start gap-0 overflow-hidden rounded-sm px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
                'hover:bg-muted/50 focus:bg-muted/50 focus:outline-none',
                isDirty ? 'bg-warning-muted dark:bg-warning-muted/60' : '',
                disabled ? 'cursor-not-allowed opacity-50' : ''
              ]
                .filter(Boolean)
                .join(' ')}
            >
              {#if type === 'number'}
                <code class="block max-w-full truncate rounded-md bg-muted px-1.5 py-0.5 text-sm">
                  {displayValue}
                </code>
              {:else}
                <span class="min-w-0 truncate">{displayValue}</span>
              {/if}
            </Button>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content side="top" class="max-w-xs">
          <p class="break-words">{displayTitle}</p>
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
        'editable-cell-display h-7 min-h-7 w-full min-w-0 shrink cursor-pointer items-center justify-start gap-0 overflow-hidden rounded-sm px-2 py-1 text-left text-sm font-normal whitespace-normal transition-colors',
        'hover:bg-muted/50 focus:bg-muted/50 focus:outline-none',
        isDirty ? 'bg-warning-muted dark:bg-warning-muted/60' : '',
        disabled ? 'cursor-not-allowed opacity-50' : ''
      ]
        .filter(Boolean)
        .join(' ')}
    >
      {#if displayValue}
        {#if type === 'number'}
          <code class="block max-w-full truncate rounded-md bg-muted px-1.5 py-0.5 text-sm">
            {displayValue}
          </code>
        {:else}
          <span class="min-w-0 truncate">{displayValue}</span>
        {/if}
      {:else}
        <span class="min-w-0 truncate text-muted-foreground">{emptyText}</span>
      {/if}
    </Button>
    {#if canUndo}
      <InlineUndoButton title={undoTitle} onclick={() => onUndo?.()} />
    {/if}
  </div>
{/if}
