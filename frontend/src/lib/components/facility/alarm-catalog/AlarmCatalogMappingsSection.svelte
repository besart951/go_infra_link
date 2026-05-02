<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { History, Trash2 } from '@lucide/svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';
  import type { AlarmTypeField } from '$lib/domain/facility/alarm-type.js';
  import type { AlarmCatalogState } from './AlarmCatalogState.svelte.js';

  interface Props {
    state: AlarmCatalogState;
  }

  let { state: catalogState }: Props = $props();
  const t = createTranslator();
  let historyItem = $state<AlarmTypeField | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.alarm_field?.label ?? historyItem.alarm_field_id}`}
    entityTable="alarm_type_fields"
    entityId={historyItem.id}
    onRestored={() => catalogState.loadAll()}
  />
{/if}

<Card.Root>
  <Card.Header class="border-b">
    <Card.Title>{$t('facility.alarm_catalog_page.mappings.title')}</Card.Title>
    <Card.Description>{$t('facility.alarm_catalog_page.mappings.description')}</Card.Description>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="space-y-2">
      <Label for="mapping-type">{$t('facility.alarm_catalog_page.labels.alarm_type')}</Label>
      <select
        id="mapping-type"
        class={catalogState.selectClass}
        bind:value={catalogState.selectedTypeId}
        onchange={() => catalogState.selectType(catalogState.selectedTypeId)}
      >
        <option value="">{$t('facility.alarm_catalog_page.labels.select')}</option>
        {#each catalogState.types as type}
          <option value={type.id}>{type.name} ({type.code})</option>
        {/each}
      </select>
    </div>

    <div class="grid gap-3 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="mapping-field">{$t('facility.alarm_catalog_page.labels.alarm_field')}</Label>
        <select
          id="mapping-field"
          class={catalogState.selectClass}
          bind:value={catalogState.mapForm.alarm_field_id}
        >
          <option value="">{$t('facility.alarm_catalog_page.labels.select')}</option>
          {#each catalogState.fields as field}
            <option value={field.id}>{field.label} ({field.key})</option>
          {/each}
        </select>
      </div>
      <div class="space-y-2">
        <Label for="mapping-order">{$t('facility.alarm_catalog_page.labels.display_order')}</Label>
        <Input id="mapping-order" type="number" bind:value={catalogState.mapForm.display_order} />
      </div>
      <div class="space-y-2">
        <Label for="mapping-group">{$t('facility.alarm_catalog_page.labels.ui_group')}</Label>
        <Input id="mapping-group" bind:value={catalogState.mapForm.ui_group} />
      </div>
      <div class="space-y-2">
        <Label for="mapping-unit">{$t('facility.alarm_catalog_page.labels.default_unit')}</Label>
        <select
          id="mapping-unit"
          class={catalogState.selectClass}
          bind:value={catalogState.mapForm.default_unit_id}
        >
          <option value="">-</option>
          {#each catalogState.units as unit}
            <option value={unit.id}>{unit.code}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="flex flex-wrap gap-6">
      <label class="flex items-center gap-2 text-sm text-foreground">
        <Checkbox bind:checked={catalogState.mapForm.is_required} />
        {$t('facility.alarm_catalog_page.labels.required')}
      </label>
      <label class="flex items-center gap-2 text-sm text-foreground">
        <Checkbox bind:checked={catalogState.mapForm.is_user_editable} />
        {$t('facility.alarm_catalog_page.labels.editable')}
      </label>
    </div>

    <div class="flex justify-end">
      {#if canPerform('update', 'alarmtype')}
        <Button
          onclick={() => catalogState.createMapping()}
          disabled={!catalogState.selectedTypeId || !catalogState.mapForm.alarm_field_id}
        >
          {$t('facility.alarm_catalog_page.mappings.create')}
        </Button>
      {/if}
    </div>

    <div class="overflow-hidden rounded-md border">
      <div class="max-h-72 overflow-auto">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.field')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.group')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.required')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.order')}</Table.Head>
              <Table.Head class="w-24 text-right"
                >{$t('facility.alarm_catalog_page.labels.action')}</Table.Head
              >
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#if !catalogState.selectedTypeId}
              <Table.Row>
                <Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.mappings.select_type_empty')}
                </Table.Cell>
              </Table.Row>
            {:else if catalogState.typeFields.length === 0}
              <Table.Row>
                <Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.mappings.empty')}
                </Table.Cell>
              </Table.Row>
            {:else}
              {#each catalogState.typeFields as typeField}
                <Table.Row>
                  <Table.Cell class="font-medium">
                    {typeField.alarm_field?.label ?? typeField.alarm_field_id}
                  </Table.Cell>
                  <Table.Cell>{typeField.ui_group ?? '-'}</Table.Cell>
                  <Table.Cell>
                    {typeField.is_required
                      ? $t('facility.alarm_catalog_page.labels.yes')
                      : $t('facility.alarm_catalog_page.labels.no')}
                  </Table.Cell>
                  <Table.Cell>{typeField.display_order}</Table.Cell>
                  <Table.Cell class="text-right">
                    <div class="flex justify-end gap-1">
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        onclick={() => {
                          historyItem = typeField;
                          historyOpen = true;
                        }}
                        aria-label={$t('history.open')}
                        title={$t('history.open')}
                      >
                        <History class="size-4" />
                      </Button>
                      {#if canPerform('update', 'alarmtype')}
                        <Button
                          size="icon-sm"
                          variant="ghost"
                          class="text-destructive hover:text-destructive"
                          onclick={() => catalogState.deleteMapping(typeField.id)}
                          aria-label={$t('facility.alarm_catalog_page.mappings.delete')}
                          title={$t('facility.alarm_catalog_page.mappings.delete')}
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
