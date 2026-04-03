<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { cn } from '$lib/utils.js';

  export interface SMTPModeOption {
    value: string;
    label: string;
    description: string;
    badge?: string;
  }

  interface Props {
    value?: string;
    label: string;
    description?: string;
    options: SMTPModeOption[];
    columnsClass?: string;
    onChange?: (value: string) => void;
    class?: string;
  }

  let {
    value = $bindable(''),
    label,
    description,
    options,
    columnsClass = 'md:grid-cols-3',
    onChange,
    class: className
  }: Props = $props();

  function selectOption(nextValue: string) {
    value = nextValue;
    onChange?.(nextValue);
  }
</script>

<div class={cn('space-y-3', className)}>
  <div class="space-y-1">
    <p class="text-sm font-medium">{label}</p>
    {#if description}
      <p class="text-sm leading-6 text-muted-foreground">{description}</p>
    {/if}
  </div>

  <div role="radiogroup" aria-label={label} class={cn('grid gap-3', columnsClass)}>
    {#each options as option (option.value)}
      <button
        type="button"
        role="radio"
        aria-checked={value === option.value}
        class={cn(
          'rounded-xl border bg-background p-4 text-left transition-colors outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50',
          value === option.value ? 'border-primary bg-primary/5 shadow-sm' : 'hover:bg-accent/60'
        )}
        onclick={() => selectOption(option.value)}
      >
        <div class="flex items-start justify-between gap-3">
          <span class="font-medium">{option.label}</span>
          {#if option.badge}
            <Badge variant={value === option.value ? 'default' : 'outline'}>{option.badge}</Badge>
          {/if}
        </div>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">{option.description}</p>
      </button>
    {/each}
  </div>
</div>
