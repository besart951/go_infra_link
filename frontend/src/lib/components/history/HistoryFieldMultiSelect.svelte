<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Command from '$lib/components/ui/command/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { cn } from '$lib/utils.js';
  import CheckIcon from '@lucide/svelte/icons/check';
  import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
  import XIcon from '@lucide/svelte/icons/x';

  interface FieldOption {
    id: string;
    label: string;
  }

  interface Props {
    items: FieldOption[];
    value?: string[];
    disabled?: boolean;
    placeholder: string;
    searchPlaceholder: string;
    emptyText: string;
    selectedText: string;
    width?: string;
    showSelectedBadges?: boolean;
  }

  let {
    items,
    value = $bindable([]),
    disabled = false,
    placeholder,
    searchPlaceholder,
    emptyText,
    selectedText,
    width = 'w-full',
    showSelectedBadges = true
  }: Props = $props();

  let open = $state(false);
  let search = $state('');

  const selectedItems = $derived(items.filter((item) => value.includes(item.id)));
  const filteredItems = $derived(
    search ? items.filter((item) => item.label.toLowerCase().includes(search.toLowerCase())) : items
  );

  $effect(() => {
    const validIds = new Set(items.map((item) => item.id));
    const next = value.filter((id) => validIds.has(id));
    if (next.length !== value.length) {
      value = next;
    }
  });

  function toggle(id: string): void {
    value = value.includes(id) ? value.filter((item) => item !== id) : [...value, id];
  }

  function remove(id: string): void {
    value = value.filter((item) => item !== id);
  }
</script>

<div class={cn('space-y-2', width)}>
  {#if showSelectedBadges && selectedItems.length > 0}
    <div class="flex flex-wrap gap-1.5">
      {#each selectedItems as item (item.id)}
        <Badge variant="secondary" class="max-w-48 gap-1 pr-1 pl-2">
          <span class="truncate">{item.label}</span>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            class="size-4 rounded-full p-0 hover:bg-secondary-foreground/20"
            onclick={() => remove(item.id)}
            {disabled}
            aria-label={item.label}
          >
            <XIcon class="size-3" />
          </Button>
        </Badge>
      {/each}
    </div>
  {/if}

  <Popover.Root bind:open>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button
          {...props}
          variant="outline"
          role="combobox"
          aria-expanded={open}
          class={cn('justify-between', width)}
          {disabled}
        >
          {selectedItems.length > 0 ? selectedText : placeholder}
          <ChevronsUpDownIcon class="size-4 opacity-50" />
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content class={cn('p-0', width)}>
      <Command.Root shouldFilter={false}>
        <Command.Input placeholder={searchPlaceholder} bind:value={search} />
        <Command.List>
          <Command.Empty>{emptyText}</Command.Empty>
          <Command.Group>
            {#each filteredItems as item (item.id)}
              <Command.Item value={item.id} onSelect={() => toggle(item.id)}>
                <CheckIcon
                  class={cn('mr-2 size-4', value.includes(item.id) ? 'opacity-100' : 'opacity-0')}
                />
                {item.label}
              </Command.Item>
            {/each}
          </Command.Group>
        </Command.List>
      </Command.Root>
    </Popover.Content>
  </Popover.Root>
</div>
