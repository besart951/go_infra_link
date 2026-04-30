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
    <Card.Title>{$t('facility.alarm_catalog_page.types.title')}</Card.Title>
    <Card.Description>{$t('facility.alarm_catalog_page.types.description')}</Card.Description>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="grid gap-3 md:grid-cols-2">
      <div class="space-y-2">
        <Label for="type-code">{$t('facility.alarm_catalog_page.labels.code')}</Label>
        <Input id="type-code" bind:value={state.typeForm.code} />
      </div>
      <div class="space-y-2">
        <Label for="type-name">{$t('common.name')}</Label>
        <Input id="type-name" bind:value={state.typeForm.name} />
      </div>
    </div>
    <div class="flex justify-end">
      {#if canPerform('create', 'alarmtype')}
        <Button
          onclick={() => state.createType()}
          disabled={!state.typeForm.code || !state.typeForm.name}
        >
          {$t('facility.alarm_catalog_page.types.create')}
        </Button>
      {/if}
    </div>
    <div class="overflow-hidden rounded-md border">
      <div class="max-h-72 overflow-auto">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>{$t('facility.alarm_catalog_page.labels.code')}</Table.Head>
              <Table.Head>{$t('common.name')}</Table.Head>
              <Table.Head class="w-24 text-right"
                >{$t('facility.alarm_catalog_page.labels.action')}</Table.Head
              >
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#if state.types.length === 0}
              <Table.Row>
                <Table.Cell colspan={3} class="py-8 text-center text-sm text-muted-foreground">
                  {$t('facility.alarm_catalog_page.types.empty')}
                </Table.Cell>
              </Table.Row>
            {:else}
              {#each state.types as type}
                <Table.Row>
                  <Table.Cell class="font-medium">{type.code}</Table.Cell>
                  <Table.Cell>{type.name}</Table.Cell>
                  <Table.Cell class="text-right">
                    {#if canPerform('delete', 'alarmtype')}
                      <Button
                        size="icon-sm"
                        variant="ghost"
                        class="text-destructive hover:text-destructive"
                        onclick={() => state.deleteType(type.id)}
                        aria-label={$t('facility.alarm_catalog_page.types.delete')}
                        title={$t('facility.alarm_catalog_page.types.delete')}
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
