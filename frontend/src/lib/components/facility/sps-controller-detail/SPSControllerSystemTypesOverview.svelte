<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import SearchIcon from '@lucide/svelte/icons/search';
  import { useSPSControllerDetailState } from './state/context.svelte.js';

  const state = useSPSControllerDetailState();
  const t = createTranslator();
</script>

<Card.Root class="border-border/70 bg-card/80">
  <Card.Header>
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1.5">
        <Card.Title class="text-xl"
          >{$t('facility.sps_controller_detail.system_types_title')}</Card.Title
        >
        <Card.Description>{$t('facility.sps_controller_detail.system_types_desc')}</Card.Description
        >
      </div>

      <Badge variant="secondary">{state.systemTypeCountLabel}</Badge>
    </div>
  </Card.Header>

  <Card.Content>
    <div class="relative mb-4">
      <SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        value={state.systemTypeSearchQuery}
        placeholder={$t('facility.sps_controller_detail.search_system_types')}
        class="pl-9"
        oninput={(event) =>
          state.setSystemTypeSearchQuery((event.currentTarget as HTMLInputElement).value)}
      />
    </div>

    {#if state.systemTypeRows.length === 0}
      <div
        class="rounded-md border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground"
      >
        {#if state.systemTypeSearchQuery.trim()}
          {$t('facility.no_items')}
        {:else}
          {$t('facility.sps_controller_detail.no_system_types')}
        {/if}
      </div>
    {:else}
      <div class="overflow-hidden rounded-lg border border-border/70">
        <Table.Root>
          <Table.Header class="bg-muted/20">
            <Table.Row>
              <Table.Head class="w-[140px]"
                >{$t('facility.control_cabinet_detail.number')}</Table.Head
              >
              <Table.Head>{$t('facility.sps_controller_detail.document_name')}</Table.Head>
              <Table.Head class="w-[120px]">{$t('common.view')}</Table.Head>
              <Table.Head class="w-[140px] text-right"
                >{$t('facility.sps_controller_detail.field_devices')}</Table.Head
              >
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {#each state.systemTypeRows as row (row.id)}
              <Table.Row>
                <Table.Cell class="font-medium text-foreground">
                  <a href={state.getSystemTypeHref(row.id)} class="hover:underline">{row.number}</a>
                </Table.Cell>
                <Table.Cell>
                  <div class="space-y-1">
                    <div class="font-medium text-foreground">
                      <a href={state.getSystemTypeHref(row.id)} class="hover:underline">
                        {row.documentName}
                      </a>
                    </div>
                    {#if row.systemTypeName}
                      <div class="text-xs text-muted-foreground">{row.systemTypeName}</div>
                    {/if}
                  </div>
                </Table.Cell>
                <Table.Cell>
                  <a
                    href={state.getSystemTypeActionHref(row.id)}
                    class="text-sm font-medium text-primary hover:underline"
                  >
                    {$t(state.canUpdateSpsControllerSystemType ? 'common.edit' : 'common.view')}
                  </a>
                </Table.Cell>
                <Table.Cell class="text-right">
                  <Badge variant="outline" class="min-w-12 justify-center">
                    {row.fieldDevicesCount}
                  </Badge>
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      </div>
    {/if}
  </Card.Content>
</Card.Root>
