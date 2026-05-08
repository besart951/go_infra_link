<script lang="ts" generics="T">
  import * as Command from '$lib/components/ui/command/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { cn } from '$lib/utils.js';
  import { Check, ChevronsUpDown } from '@lucide/svelte';

  interface StaticComboboxProps<T> {
    items: T[];
    value?: string;
    labelKey: keyof T;
    optionLabelKey?: keyof T;
    tooltipLabelKey?: keyof T;
    idKey?: keyof T;
    id?: string;
    disabled?: boolean;
    clearable?: boolean;
    clearText?: string;
    placeholder?: string;
    searchPlaceholder?: string;
    emptyText?: string;
    width?: string;
    popupWidth?: string;
    onValueChange?: (value: string) => void;
    error?: string;
  }

  let {
    items,
    value = $bindable(''),
    labelKey,
    optionLabelKey = labelKey,
    tooltipLabelKey,
    idKey = 'id' as keyof T,
    id,
    disabled = false,
    clearable = false,
    clearText = 'Clear selection',
    placeholder = 'Select item...',
    searchPlaceholder = 'Search...',
    emptyText = 'No results found.',
    width = 'w-[200px]',
    popupWidth = width,
    onValueChange,
    error
  }: StaticComboboxProps<T> = $props();

  let open = $state(false);
  let search = $state('');

  const selectedItem = $derived(items.find((i) => String(i[idKey]) === value));
  const selectedLabel = $derived(selectedItem ? String(selectedItem[labelKey] ?? '') : undefined);
  const selectedTooltipLabel = $derived(
    selectedItem && tooltipLabelKey ? String(selectedItem[tooltipLabelKey] ?? '') : undefined
  );
  const triggerLabel = $derived(selectedLabel || (value ? value : placeholder));
  const triggerTitle = $derived(
    selectedTooltipLabel || selectedLabel || (value ? value : undefined)
  );
  const hasError = $derived(!!error);

  const filteredItems = $derived(
    search
      ? items.filter((i) =>
          [
            String(i[labelKey] ?? ''),
            String(i[optionLabelKey] ?? ''),
            tooltipLabelKey ? String(i[tooltipLabelKey] ?? '') : ''
          ]
            .join(' ')
            .toLowerCase()
            .includes(search.toLowerCase())
        )
      : items
  );

  function clearSelection() {
    value = '';
    onValueChange?.('');
    open = false;
  }

  function mergeTriggerProps(
    popoverProps: Record<string, unknown>,
    tooltipProps: Record<string, unknown>
  ): Record<string, unknown> {
    const merged: Record<string, unknown> = { ...tooltipProps, ...popoverProps };
    for (const key of new Set([...Object.keys(tooltipProps), ...Object.keys(popoverProps)])) {
      const tooltipHandler = tooltipProps[key];
      const popoverHandler = popoverProps[key];
      if (
        !key.startsWith('on') ||
        typeof tooltipHandler !== 'function' ||
        typeof popoverHandler !== 'function'
      ) {
        continue;
      }
      merged[key] = (...args: unknown[]) => {
        tooltipHandler(...args);
        popoverHandler(...args);
      };
    }
    return merged;
  }
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      {#if hasError}
        <Tooltip.Provider>
          <Tooltip.Root>
            <Tooltip.Trigger>
              {#snippet child({ props: tooltipProps })}
                <Button
                  {...mergeTriggerProps(props, tooltipProps)}
                  {id}
                  variant="outline"
                  role="combobox"
                  aria-expanded={open}
                  {disabled}
                  aria-disabled={disabled}
                  class={cn(
                    'min-w-0 justify-between overflow-hidden border-destructive text-destructive',
                    width
                  )}
                >
                  <span class="min-w-0 truncate text-left">{triggerLabel}</span>
                  <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              {/snippet}
            </Tooltip.Trigger>
            <Tooltip.Content side="top">
              <p>{error}</p>
            </Tooltip.Content>
          </Tooltip.Root>
        </Tooltip.Provider>
      {:else if triggerTitle}
        <Tooltip.Provider>
          <Tooltip.Root>
            <Tooltip.Trigger>
              {#snippet child({ props: tooltipProps })}
                <Button
                  {...mergeTriggerProps(props, tooltipProps)}
                  {id}
                  variant="outline"
                  role="combobox"
                  aria-expanded={open}
                  {disabled}
                  aria-disabled={disabled}
                  class={cn('min-w-0 justify-between overflow-hidden', width)}
                >
                  <span class="min-w-0 truncate text-left">{triggerLabel}</span>
                  <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
                </Button>
              {/snippet}
            </Tooltip.Trigger>
            <Tooltip.Content side="top" class="max-w-xs">
              <p class="wrap-break-word">{triggerTitle}</p>
            </Tooltip.Content>
          </Tooltip.Root>
        </Tooltip.Provider>
      {:else}
        <Button
          {...props}
          {id}
          variant="outline"
          role="combobox"
          aria-expanded={open}
          {disabled}
          aria-disabled={disabled}
          class={cn('min-w-0 justify-between overflow-hidden', width)}
        >
          <span class="min-w-0 truncate text-left">{triggerLabel}</span>
          <ChevronsUpDown class="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      {/if}
    {/snippet}
  </Popover.Trigger>
  <Popover.Content class={cn('p-0', popupWidth)}>
    <Command.Root shouldFilter={false}>
      <Command.Input placeholder={searchPlaceholder} bind:value={search} />
      <Command.List>
        <Command.Empty>{emptyText}</Command.Empty>
        <Command.Group>
          {#if clearable && value}
            <Command.Item
              value=""
              onSelect={() => {
                clearSelection();
              }}
            >
              {clearText}
            </Command.Item>
          {/if}
          {#each filteredItems as item (String(item[idKey]))}
            <Command.Item
              value={String(item[idKey])}
              onSelect={() => {
                const next = String(item[idKey] ?? '');
                if (!next || next === 'undefined' || next === 'null') return;
                value = next;
                onValueChange?.(value);
                open = false;
              }}
            >
              <Check
                class={cn(
                  'mr-2 h-4 w-4',
                  value === String(item[idKey]) ? 'opacity-100' : 'opacity-0'
                )}
              />
              {String(item[optionLabelKey] ?? '')}
            </Command.Item>
          {/each}
        </Command.Group>
      </Command.List>
    </Command.Root>
  </Popover.Content>
</Popover.Root>
