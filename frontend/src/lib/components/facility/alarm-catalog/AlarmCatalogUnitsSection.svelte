<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { History, Trash2 } from '@lucide/svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';
  import type { Unit } from '$lib/domain/facility/alarm-type.js';
  import type { AlarmCatalogState } from './AlarmCatalogState.svelte.js';

  interface Props {
    state: AlarmCatalogState;
  }

  let { state: catalogState }: Props = $props();
  const t = createTranslator();
  let historyItem = $state<Unit | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.code}`}
    entityTable="units"
    entityId={historyItem.id}
    onRestored={() => catalogState.loadAll()}
  />
{/if}

<Card.Root>
  <Card.Header class="border-b">
    <Card.Title>{$t('facility.alarm_catalog_page.units.title')}</Card.Title>
    <Card.Description>{$t('facility.alarm_catalog_page.units.description')}</Card.Description>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="grid gap-3 md:grid-cols-3">
      <div class="space-y-2">
        <Label for="unit-code">{$t('facility.alarm_catalog_page.labels.code')}</Label>
        <Input id="unit-code" bind:value={catalogState.unitForm.code} />
      </div>
      <div class="space-y-2">
        <Label for="unit-symbol">{$t('facility.alarm_catalog_page.labels.symbol')}</Label>
        <Input id="unit-symbol" bind:value={catalogState.unitForm.symbol} />
      </div>
      <div class="space-y-2">
        <Label for="unit-name">{$t('common.name')}</Label>
        <Input id="unit-name" bind:value={catalogState.unitForm.name} />
      </div>
    </div>
    <div class="flex justify-end">
      {#if canPerform('create', 'alarmtype')}
        <Button
          onclick={() => catalogState.createUnit()}
          disabled={!catalogState.unitForm.code ||
            !catalogState.unitForm.symbol ||
            !catalogState.unitForm.name}
        >
          {$t('facility.alarm_catalog_page.units.create')}
        </Button>
      {/if}
    </div>
    <div class="overflow-hidden rounded-md border">
      <div class="max-h-72 overflow-auto">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.code')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.symbol')}</Table.Head>
              <Table.Head>{$t('common.name')}</Table.Head>
              <Table.Head class="w-24 text-right"
                >{$t('facility.alarm_catalog_page.labels.action')}</Table.Head
              >
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#if catalogState.units.length === 0}
              <Table.Row>
                <Table.Cell colspan={4} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.units.empty')}
                </Table.Cell>
              </Table.Row>
            {:else}
              {#each catalogState.units as unit}
                <Table.Row>
                  <Table.Cell class="font-medium">{unit.code}</Table.Cell>
                  <Table.Cell>{unit.symbol}</Table.Cell>
                  <Table.Cell>{unit.name}</Table.Cell>
                  <Table.Cell class="text-right">
                    <div class="flex justify-end gap-1">
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        onclick={() => {
                          historyItem = unit;
                          historyOpen = true;
                        }}
                        aria-label={$t('history.open')}
                        title={$t('history.open')}
                      >
                        <History class="size-4" />
                      </Button>
                      {#if canPerform('delete', 'alarmtype')}
                        <Button
                          size="icon-sm"
                          variant="ghost"
                          class="text-destructive hover:text-destructive"
                          onclick={() => catalogState.deleteUnit(unit.id)}
                          aria-label={$t('facility.alarm_catalog_page.units.delete')}
                          title={$t('facility.alarm_catalog_page.units.delete')}
                        >
                          <Trash2 class="size-4" />
                        </Button>
                      {/if}
                    </div>
                  </Table.Cell>
                </Table.Row>
              {/each}
            {/if}
          </Table.Body>
        </Table.Root>
      </div>
    </div>
  </Card.Content>
</Card.Root>
