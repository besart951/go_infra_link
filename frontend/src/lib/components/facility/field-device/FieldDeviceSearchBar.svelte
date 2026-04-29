<script lang="ts">
  import { buttonVariants } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Search, Trash2, Settings2, TableIcon, Filter, X, RefreshCcw } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useFieldDeviceState();

  function handleSearchInput(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    void state.search(value);
  }
</script>

<div class="flex items-center gap-3">
  <div class="relative flex-1">
    <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
    <Input
      type="search"
      placeholder={$t('field_device.search.placeholder')}
      class="pl-9"
      value={state.searchText}
      oninput={handleSearchInput}
    />
  </div>

  <div class="ml-auto flex items-center gap-2">
    {#if state.selectedCount > 0}
      <span class="text-sm text-muted-foreground">
        {$t('field_device.search.selected', { count: state.selectedCount })}
      </span>
    {/if}

    <Tooltip.Provider>
      <ButtonGroup.Root>
        {#if state.selectedCount > 0}
          <Tooltip.Root>
            <Tooltip.Trigger
              class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
              onclick={() => state.clearSelection()}
            >
              <X />
            </Tooltip.Trigger>
            <Tooltip.Content>{$t('field_device.search.clear')}</Tooltip.Content>
          </Tooltip.Root>

          {#if state.canDeleteFieldDevice()}
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({ variant: 'destructive', size: 'icon-sm' })}
                onclick={() => void state.bulkDeleteSelected()}
              >
                <Trash2 />
              </Tooltip.Trigger>
              <Tooltip.Content>{$t('field_device.search.delete')}</Tooltip.Content>
            </Tooltip.Root>
          {/if}

          {#if state.canOpenBulkEditPanel()}
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({
                  variant: state.showBulkEditPanel ? 'secondary' : 'outline',
                  size: 'icon-sm'
                })}
                onclick={() => state.toggleBulkEditPanel()}
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
              variant: state.showExportPanel ? 'secondary' : 'outline',
              size: 'icon-sm'
            })}
            onclick={() => state.toggleExportPanel()}
          >
            <TableIcon />
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('field_device.search.table')}</Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger
            class={`${buttonVariants({
              variant: state.showFilterPanel ? 'secondary' : 'outline',
              size: 'icon-sm'
            })} relative`}
            onclick={() => state.toggleFilterPanel()}
          >
            <Filter />
            {#if state.hasActiveFilters}
              <span
                class="pointer-events-none absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-green-500 ring-2 ring-background"
              ></span>
            {/if}
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('common.filter')}</Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
            onclick={() => void state.reload()}
            disabled={state.loading}
          >
            <RefreshCcw />
          </Tooltip.Trigger>
          <Tooltip.Content>{$t('field_device.search.refresh')}</Tooltip.Content>
        </Tooltip.Root>
      </ButtonGroup.Root>
    </Tooltip.Provider>
  </div>
</div>
