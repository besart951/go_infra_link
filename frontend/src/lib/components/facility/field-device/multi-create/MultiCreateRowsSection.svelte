<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import FieldDeviceRow from '../FieldDeviceRow.svelte';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Alert from '$lib/components/ui/alert/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import { Plus, CircleAlert } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  import type { FieldDeviceMultiCreateState } from './FieldDeviceMultiCreateState.svelte.js';

  type Props = {
    state: FieldDeviceMultiCreateState;
    onCancel?: () => void;
  };

  let { state, onCancel }: Props = $props();

  const t = createTranslator();
</script>

<div class="p-6">
  <div class="mb-4 flex justify-end">
    <Button onclick={() => state.addRow()} disabled={!state.canAddRow} size="sm">
      <Plus class="mr-2 size-4" />
      {$t('field_device.multi_create.rows.add')}
    </Button>
  </div>

  {#if state.rows.length === 0}
    <Alert.Root>
      <CircleAlert class="size-4" />
      <Alert.Description>
        {#if state.availableNumbers.length === 0 && !state.loadingAvailableNumbers}
          {$t('field_device.multi_create.rows.none_available')}
        {:else if state.loadingAvailableNumbers}
          {$t('field_device.multi_create.rows.loading_numbers')}
        {:else}
          {$t('field_device.multi_create.rows.empty_prompt')}
        {/if}
      </Alert.Description>
    </Alert.Root>
  {/if}

  {#if state.rows.length > 0}
    <div class="overflow-x-auto rounded-lg border">
      <Table.Root class="[&_td]:p-2 [&_th]:h-10 [&_th]:px-2">
        <Table.Header>
          <Table.Row class="my-0">
            <Table.Head class="text-center">#</Table.Head>
            <Table.Head>{$t('field_device.row.bmk')}</Table.Head>
            <Table.Head>{$t('field_device.row.description')}</Table.Head>
            <Table.Head>{$t('field_device.row.text_fix')}</Table.Head>
            <Table.Head>{$t('field_device.row.apparat_nr')}</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#each state.rows as row, index (row.id)}
            <FieldDeviceRow {state} {index} />
          {/each}
        </Table.Body>
      </Table.Root>
    </div>

    <Separator class="my-4" />

    <div class="flex items-center justify-between">
      <p class="text-sm text-muted-foreground">
        {$t('field_device.multi_create.rows.summary', { count: state.rows.length })}
        {#if state.hasValidationErrors}
          <span class="text-destructive">
            {$t('field_device.multi_create.rows.errors', { count: state.rowErrors.size })}
          </span>
        {/if}
      </p>
      <div class="flex gap-2">
        {#if onCancel}
          <Button variant="outline" onclick={onCancel} disabled={state.submitting}>
            {$t('common.cancel')}
          </Button>
        {/if}
        <Button
          onclick={() => state.handleSubmit()}
          disabled={state.submitting || state.rows.length === 0 || state.hasValidationErrors}
        >
          {state.submitting
            ? $t('field_device.multi_create.actions.creating')
            : $t('field_device.multi_create.actions.create', { count: state.rows.length })}
        </Button>
      </div>
    </div>
  {/if}
</div>
