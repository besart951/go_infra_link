<script lang="ts">
  import { Button, type ButtonProps, type ButtonSize } from '$lib/components/ui/button/index.js';
  import { cn } from '$lib/utils.js';
  import type { Snippet } from 'svelte';

  type ToggleButtonProps = Omit<ButtonProps, 'variant' | 'size'> & {
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

<Button
  bind:ref
  {type}
  variant={pressed ? 'secondary' : 'outline'}
  {size}
  aria-pressed={pressed}
  data-slot="toggle-button"
  data-state={pressed ? 'on' : 'off'}
  class={cn(
    'justify-start text-left whitespace-normal',
    pressed && 'border-primary/40 bg-primary/10 text-primary shadow-sm hover:bg-primary/15',
    className
  )}
  {...restProps}
>
  {@render children?.()}
</Button>
