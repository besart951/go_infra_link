<script lang="ts">
  import { onDestroy } from 'svelte';
  import { buttonVariants } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Search, Trash2, Settings2, TableIcon, Filter, X, RefreshCcw } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import FieldDeviceViewPopover from './FieldDeviceViewPopover.svelte';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const fieldDeviceState = useFieldDeviceState();
  const SEARCH_DEBOUNCE_MS = 300;

  let searchText = $state(fieldDeviceState.searchText);
  let searchDebounceTimer: ReturnType<typeof setTimeout> | undefined;

  function handleSearchInput(event: Event) {
    searchText = (event.target as HTMLInputElement).value;
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(() => {
      void fieldDeviceState.search(searchText);
    }, SEARCH_DEBOUNCE_MS);
  }

  $effect(() => {
    searchText = fieldDeviceState.searchText;
  });

  onDestroy(() => {
    clearTimeout(searchDebounceTimer);
  });
</script>

<div class="flex items-center gap-3">
  <div class="relative flex-1">
    <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    <Input
      type="search"
      placeholder={$t('field_device.search.placeholder')}
      class="pl-9"
      value={searchText}
      oninput={handleSearchInput}
    />
  </div>

  <div class="ml-auto flex items-center gap-2">
    {#if fieldDeviceState.selectedCount > 0}
      <span class="text-sm text-muted-foreground">
        {$t('field_device.search.selected', { count: fieldDeviceState.selectedCount })}
      </span>
    {/if}

    <Tooltip.Provider>
      <ButtonGroup.Root>
        {#if fieldDeviceState.selectedCount > 0}
          <Tooltip.Root>
            <Tooltip.Trigger
              class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
              onclick={() => fieldDeviceState.clearSelection()}
            >
              <X />
            </Tooltip.Trigger>
            <Tooltip.Content>{$t('field_device.search.clear')}</Tooltip.Content>
          </Tooltip.Root>

          {#if fieldDeviceState.canDeleteFieldDevice()}
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({ variant: 'destructive', size: 'icon-sm' })}
                onclick={() => void fieldDeviceState.bulkDeleteSelected()}
              >
                <Trash2 />
              </Tooltip.Trigger>
              <Tooltip.Content>{$t('field_device.search.delete')}</Tooltip.Content>
            </Tooltip.Root>
          {/if}

          {#if fieldDeviceState.canOpenBulkEditPanel()}
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({
                  variant: fieldDeviceState.showBulkEditPanel ? 'secondary' : 'outline',
                  size: 'icon-sm'
                })}
                onclick={() => fieldDeviceState.toggleBulkEditPanel()}
              >
                <Settings2 />
              </Tooltip.Trigger>
              <Tooltip.Content>{$t('field_device.search.bulk_edit')}</Tooltip.Content>
            </Tooltip.Root>
          {/if}
        {/if}

        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({
              variant: fieldDeviceState.showExportPanel ? 'secondary' : 'outline',
              size: 'icon-sm'
            })}
            onclick={() => fieldDeviceState.toggleExportPanel()}
          >
            <TableIcon />
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('field_device.search.table')}</Tooltip.Content>
        </Tooltip.Root>

        <FieldDeviceViewPopover />

        <Tooltip.Root>
          <Tooltip.Trigger
            class={`${buttonVariants({
              variant: fieldDeviceState.showFilterPanel ? 'secondary' : 'outline',
              size: 'icon-sm'
            })} relative`}
            onclick={() => fieldDeviceState.toggleFilterPanel()}
          >
            <Filter />
            {#if fieldDeviceState.hasActiveFilters}
              <span
                class="pointer-events-none absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-success ring-2 ring-background"
              ></span>
            {/if}
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('common.filter')}</Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
            onclick={() => void fieldDeviceState.reload()}
            disabled={fieldDeviceState.loading}
          >
            <RefreshCcw />
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('field_device.search.refresh')}</Tooltip.Content>
        </Tooltip.Root>
      </ButtonGroup.Root>
    </Tooltip.Provider>
  </div>
</div>
