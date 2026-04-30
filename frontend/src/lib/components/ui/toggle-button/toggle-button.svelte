<script lang="ts">
  import { buttonVariants, type ButtonSize } from '$lib/components/ui/button/index.js';
  import { cn, type WithElementRef } from '$lib/utils.js';
  import type { Snippet } from 'svelte';
  import type { HTMLButtonAttributes } from 'svelte/elements';

  type ToggleButtonProps = WithElementRef<HTMLButtonAttributes> & {
    pressed?: boolean;
    size?: ButtonSize;
    children?: Snippet;
  };

  let {
    ref = $bindable(null),
    class: className,
    pressed = false,
    size = 'sm',
    type = 'button',
    children,
    ...restProps
  }: ToggleButtonProps = $props();
</script>

<button
  bind:this={ref}
  {type}
  aria-pressed={pressed}
  data-slot="toggle-button"
  data-state={pressed ? 'on' : 'off'}
  class={cn(
    buttonVariants({ variant: pressed ? 'secondary' : 'outline', size }),
    'justify-start text-left whitespace-normal',
    pressed && 'border-primary/40 bg-primary/10 text-primary shadow-sm hover:bg-primary/15',
    className
  )}
  {...restProps}
>
  {@render children?.()}
</button>
