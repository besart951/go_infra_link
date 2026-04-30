<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Trash2 } from '@lucide/svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';
  import type { AlarmCatalogState } from './AlarmCatalogState.svelte.js';

  interface Props {
    state: AlarmCatalogState;
  }

  let { state }: Props = $props();
  const t = createTranslator();
</script>

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
        class={state.selectClass}
        bind:value={state.selectedTypeId}
        onchange={() => state.selectType(state.selectedTypeId)}
      >
        <option value="">{$t('facility.alarm_catalog_page.labels.select')}</option>
        {#each state.types as type}
          <option value={type.id}>{type.name} ({type.code})</option>
        {/each}
      </select>
    </div>

    <div class="grid gap-3 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="mapping-field">{$t('facility.alarm_catalog_page.labels.alarm_field')}</Label>
        <select
          id="mapping-field"
          class={state.selectClass}
          bind:value={state.mapForm.alarm_field_id}
        >
          <option value="">{$t('facility.alarm_catalog_page.labels.select')}</option>
          {#each state.fields as field}
            <option value={field.id}>{field.label} ({field.key})</option>
          {/each}
        </select>
      </div>
      <div class="space-y-2">
        <Label for="mapping-order">{$t('facility.alarm_catalog_page.labels.display_order')}</Label>
        <Input id="mapping-order" type="number" bind:value={state.mapForm.display_order} />
      </div>
      <div class="space-y-2">
        <Label for="mapping-group">{$t('facility.alarm_catalog_page.labels.ui_group')}</Label>
        <Input id="mapping-group" bind:value={state.mapForm.ui_group} />
      </div>
      <div class="space-y-2">
        <Label for="mapping-unit">{$t('facility.alarm_catalog_page.labels.default_unit')}</Label>
        <select
          id="mapping-unit"
          class={state.selectClass}
          bind:value={state.mapForm.default_unit_id}
        >
          <option value="">-</option>
          {#each state.units as unit}
            <option value={unit.id}>{unit.code}</option>
          {/each}
        </select>
      </div>
    </div>

    <div class="flex flex-wrap gap-6">
      <label class="flex items-center gap-2 text-sm text-foreground">
        <Checkbox bind:checked={state.mapForm.is_required} />
        {$t('facility.alarm_catalog_page.labels.required')}
      </label>
      <label class="flex items-center gap-2 text-sm text-foreground">
        <Checkbox bind:checked={state.mapForm.is_user_editable} />
        {$t('facility.alarm_catalog_page.labels.editable')}
      </label>
    </div>

    <div class="flex justify-end">
      {#if canPerform('update', 'alarmtype')}
        <Button
          onclick={() => state.createMapping()}
          disabled={!state.selectedTypeId || !state.mapForm.alarm_field_id}
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
            {#if !state.selectedTypeId}
              <Table.Row>
                <Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.mappings.select_type_empty')}
                </Table.Cell>
              </Table.Row>
            {:else if state.typeFields.length === 0}
              <Table.Row>
                <Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.mappings.empty')}
                </Table.Cell>
              </Table.Row>
            {:else}
              {#each state.typeFields as typeField}
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
                    {#if canPerform('update', 'alarmtype')}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        class="text-destructive hover:text-destructive"
                        onclick={() => state.deleteMapping(typeField.id)}
                        aria-label={$t('facility.alarm_catalog_page.mappings.delete')}
                        title={$t('facility.alarm_catalog_page.mappings.delete')}
                      >
                        <Trash2 class="size-4" />
                      </Button>
                    {/if}
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
