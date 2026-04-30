<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
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
    <Card.Title>{$t('facility.alarm_catalog_page.fields.title')}</Card.Title>
    <Card.Description>{$t('facility.alarm_catalog_page.fields.description')}</Card.Description>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="grid gap-3 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="field-key">{$t('facility.alarm_catalog_page.labels.key')}</Label>
        <Input id="field-key" bind:value={state.fieldForm.key} />
      </div>
      <div class="space-y-2">
        <Label for="field-label">{$t('facility.alarm_catalog_page.labels.label')}</Label>
        <Input id="field-label" bind:value={state.fieldForm.label} />
      </div>
      <div class="space-y-2">
        <Label for="field-datatype">{$t('facility.alarm_catalog_page.labels.data_type')}</Label>
        <select
          id="field-datatype"
          class={state.selectClass}
          bind:value={state.fieldForm.data_type}
        >
          {#each state.dataTypeOptions as dataType}
            <option value={dataType}>{dataType}</option>
          {/each}
        </select>
      </div>
      <div class="space-y-2">
        <Label for="field-unit">{$t('facility.alarm_catalog_page.labels.default_unit_code')}</Label>
        <select
          id="field-unit"
          class={state.selectClass}
          bind:value={state.fieldForm.default_unit_code}
        >
          <option value="">-</option>
          {#each state.units as unit}
            <option value={unit.code}>{unit.code}</option>
          {/each}
        </select>
      </div>
    </div>
    <div class="flex justify-end">
      {#if canPerform('create', 'alarmtype')}
        <Button
          onclick={() => state.createField()}
          disabled={!state.fieldForm.key || !state.fieldForm.label}
        >
          {$t('facility.alarm_catalog_page.fields.create')}
        </Button>
      {/if}
    </div>
    <div class="overflow-hidden rounded-md border">
      <div class="max-h-72 overflow-auto">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.key')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.label')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.type')}</Table.Head>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.unit')}</Table.Head>
              <Table.Head class="w-24 text-right"
                >{$t('facility.alarm_catalog_page.labels.action')}</Table.Head
              >
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#if state.fields.length === 0}
              <Table.Row>
                <Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.fields.empty')}
                </Table.Cell>
              </Table.Row>
            {:else}
              {#each state.fields as field}
                <Table.Row>
                  <Table.Cell class="font-medium">{field.key}</Table.Cell>
                  <Table.Cell>{field.label}</Table.Cell>
                  <Table.Cell>{field.data_type}</Table.Cell>
                  <Table.Cell>{field.default_unit_code ?? '-'}</Table.Cell>
                  <Table.Cell class="text-right">
                    {#if canPerform('delete', 'alarmtype')}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        class="text-destructive hover:text-destructive"
                        onclick={() => state.deleteField(field.id)}
                        aria-label={$t('facility.alarm_catalog_page.fields.delete')}
                        title={$t('facility.alarm_catalog_page.fields.delete')}
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
