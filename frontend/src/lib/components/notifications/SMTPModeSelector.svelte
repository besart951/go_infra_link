<script lang="ts">
  import CheckIcon from '@lucide/svelte/icons/check';
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
      {@const isSelected = value === option.value}
      <button
        type="button"
        role="radio"
        aria-checked={isSelected}
        class={cn(
          'grid h-full min-h-28 grid-cols-[1.25rem_minmax(0,1fr)] gap-3 rounded-lg border bg-background p-3 text-left transition-colors outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 sm:p-4',
          isSelected ? 'border-primary bg-primary/5 shadow-sm' : 'hover:bg-accent/60'
        )}
        onclick={() => selectOption(option.value)}
      >
        <span
          class={cn(
            'mt-0.5 flex size-5 shrink-0 items-center justify-center rounded-full border',
            isSelected
              ? 'border-primary bg-primary text-primary-foreground'
              : 'border-muted-foreground/40'
          )}
          aria-hidden="true"
        >
          {#if isSelected}
            <CheckIcon class="size-3.5" />
          {/if}
        </span>

        <div class="min-w-0 space-y-2">
          <div class="flex min-w-0 flex-wrap items-start gap-2">
            <span class="min-w-0 flex-1 text-sm leading-5 font-medium [overflow-wrap:anywhere]">
              {option.label}
            </span>
            {#if option.badge}
              <Badge
                variant={isSelected ? 'default' : 'outline'}
                class="shrink-0 text-[11px] leading-none"
              >
                {option.badge}
              </Badge>
            {/if}
          </div>

          <p class="text-sm leading-6 text-muted-foreground">{option.description}</p>
        </div>
      </button>
    {/each}
  </div>
</div>
