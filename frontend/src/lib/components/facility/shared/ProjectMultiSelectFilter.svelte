<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Command from '$lib/components/ui/command/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { cn } from '$lib/utils.js';
  import CheckIcon from '@lucide/svelte/icons/check';
  import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
  import XIcon from '@lucide/svelte/icons/x';
  import type { MultiFilterOption } from './projectFacilityListFilters.js';

  interface Props {
    items: MultiFilterOption[];
    value?: string[];
    label: string;
    placeholder: string;
    searchPlaceholder: string;
    emptyText: string;
    selectedText: string;
    clearText: string;
    width?: string;
    popupWidth?: string;
    disabled?: boolean;
    onValueChange?: (value: string[]) => void;
  }

  let {
    items,
    value = [],
    label,
    placeholder,
    searchPlaceholder,
    emptyText,
    selectedText,
    clearText,
    width = 'w-full sm:w-72',
    popupWidth = width,
    disabled = false,
    onValueChange
  }: Props = $props();

  let open = $state(false);
  let search = $state('');

  const selectedItems = $derived(items.filter((item) => value.includes(item.id)));
  const filteredItems = $derived.by(() => {
    const query = search.trim().toLowerCase();
    if (!query) return items;

    return items.filter((item) => item.label.toLowerCase().includes(query));
  });

  function setValue(nextValue: string[]): void {
    onValueChange?.(nextValue);
  }

  function toggle(id: string): void {
    setValue(value.includes(id) ? value.filter((item) => item !== id) : [...value, id]);
  }

  function remove(id: string): void {
    setValue(value.filter((item) => item !== id));
  }
</script>

<div class={cn('min-w-0 space-y-2', width)}>
  <div class="flex items-center justify-between gap-2">
    <span class="text-sm font-medium">{label}</span>
    {#if value.length > 0}
      <Button
        type="button"
        variant="ghost"
        size="sm"
        class="h-7 px-2 text-xs"
        onclick={() => setValue([])}
        {disabled}
      >
        <XIcon class="mr-1 size-3" />
        {clearText}
      </Button>
    {/if}
  </div>

  {#if selectedItems.length > 0}
    <div class="flex flex-wrap gap-1.5">
      {#each selectedItems as item (item.id)}
        <Badge variant="secondary" class="max-w-56 gap-1 pr-1 pl-2">
          <span class="truncate">{item.label}</span>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            class="size-4 rounded-full p-0 hover:bg-secondary-foreground/20"
            onclick={() => remove(item.id)}
            aria-label={item.label}
            {disabled}
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
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          class={cn('min-w-0 justify-between', width)}
          disabled={disabled || items.length === 0}
        >
          <span class="min-w-0 flex-1 truncate text-left">
            {selectedItems.length > 0 ? selectedText : placeholder}
          </span>
          <ChevronsUpDownIcon class="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content class={cn('p-0', popupWidth)}>
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
                <span class="min-w-0 flex-1 truncate">{item.label}</span>
                {#if item.count !== undefined}
                  <span class="ml-2 text-xs text-muted-foreground">{item.count}</span>
                {/if}
              </Command.Item>
            {/each}
          </Command.Group>
        </Command.List>
      </Command.Root>
    </Popover.Content>
  </Popover.Root>
</div>
