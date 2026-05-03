<script lang="ts">
  /**
   * EditableBooleanCell Component
   * Inline checkbox for boolean toggling in table cells
   */
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import InlineUndoButton from './InlineUndoButton.svelte';

  interface Props {
    value: boolean;
    pendingValue?: boolean;
    isDirty?: boolean;
    error?: string;
    disabled?: boolean;
    undoTitle?: string;
    onToggle: (value: boolean) => void;
    onUndo?: () => void;
  }

  let {
    value,
    pendingValue,
    isDirty = false,
    error,
    disabled = false,
    undoTitle = 'Undo field change',
    onToggle,
    onUndo
  }: Props = $props();

  const displayValue = $derived(pendingValue !== undefined ? pendingValue : value);
  const hasError = $derived(!!error);
  const canUndo = $derived(isDirty && !!onUndo);

  function handleChange(checked: boolean | 'indeterminate') {
    if (checked === 'indeterminate') return;
    onToggle(checked);
  }
</script>

<div
  class={[
    'group/undo relative flex items-center justify-center rounded-sm px-2 py-1',
    isDirty ? 'bg-warning-muted dark:bg-warning-muted/60' : '',
    hasError ? 'bg-destructive/10' : ''
  ]
    .filter(Boolean)
    .join(' ')}
>
  {#if hasError}
    <Tooltip.Provider>
      <Tooltip.Root>
        <Tooltip.Trigger>
          {#snippet child({ props })}
            <div {...props}>
              <Checkbox checked={displayValue} onCheckedChange={handleChange} {disabled} />
            </div>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content side="top" class="max-w-xs bg-destructive text-destructive-foreground">
          <p>{error}</p>
        </Tooltip.Content>
      </Tooltip.Root>
    </Tooltip.Provider>
  {:else}
    <Checkbox checked={displayValue} onCheckedChange={handleChange} {disabled} />
  {/if}
  {#if canUndo}
    <InlineUndoButton title={undoTitle} onclick={() => onUndo?.()} />
  {/if}
</div>
