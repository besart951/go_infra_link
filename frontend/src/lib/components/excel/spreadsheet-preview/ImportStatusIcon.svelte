<script lang="ts">
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { CircleCheck, CircleX, Info, ShieldAlert } from '@lucide/svelte';
  import {
    importStatusIconClass,
    type ImportStatusVisualKind
  } from './fieldDeviceImportPresentation.js';

  interface Props {
    kind: ImportStatusVisualKind;
    messages?: string | readonly string[];
    size?: 'sm' | 'md';
    class?: string;
  }

  let { kind, messages = [], size = 'md', class: className = '' }: Props = $props();

  const normalizedMessages = $derived(
    Array.isArray(messages) ? messages.filter(Boolean) : messages ? [messages] : []
  );
  const iconSizeClass = $derived(size === 'sm' ? 'size-3.5' : 'size-4');
  const iconClass = $derived(`${iconSizeClass} ${importStatusIconClass(kind)} ${className}`);
</script>

{#snippet icon()}
  {#if kind === 'failed'}
    <CircleX class={iconClass} />
  {:else if kind === 'delta'}
    <ShieldAlert class={iconClass} />
  {:else if kind === 'success'}
    <CircleCheck class={iconClass} />
  {:else if kind === 'identical'}
    <Info class={iconClass} />
  {/if}
{/snippet}

{#if kind !== 'none'}
  {#if normalizedMessages.length > 0}
    <Tooltip.Root>
      <Tooltip.Trigger class="inline-flex shrink-0">
        {@render icon()}
      </Tooltip.Trigger>
      <Tooltip.Content class="max-w-xs">
        <div class="space-y-1">
          {#each normalizedMessages as message}
            <div>{message}</div>
          {/each}
        </div>
      </Tooltip.Content>
    </Tooltip.Root>
  {:else}
    <span class="inline-flex shrink-0">
      {@render icon()}
    </span>
  {/if}
{/if}
