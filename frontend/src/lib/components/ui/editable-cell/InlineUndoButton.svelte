<script lang="ts">
  import { Undo2 } from '@lucide/svelte';
  import Button from '../button/button.svelte';

  interface Props {
    title: string;
    onclick: () => void;
  }

  let { title, onclick }: Props = $props();

  function handlePointer(event: Event): void {
    event.preventDefault();
    event.stopPropagation();
  }

  function handleClick(event: MouseEvent): void {
    event.preventDefault();
    event.stopPropagation();
    onclick();
  }
</script>

<Button
  type="button"
  variant="ghost"
  size="icon"
  aria-label={title}
  {title}
  class="absolute top-0 right-0 z-20 h-4 min-h-0 w-4 rounded-full border bg-background p-0 text-muted-foreground opacity-0 shadow-sm transition-opacity group-hover/undo:opacity-100 hover:text-foreground focus-visible:opacity-100"
  onpointerdown={handlePointer}
  onclick={handleClick}
>
  <Undo2 class="h-2.5 w-2.5 stroke-[2.5]" />
</Button>
