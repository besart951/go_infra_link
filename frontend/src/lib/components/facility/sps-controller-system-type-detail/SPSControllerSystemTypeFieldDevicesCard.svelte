<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import type { FieldDevice } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    fieldDevices: FieldDevice[];
    total: number;
  }

  let { fieldDevices, total }: Props = $props();

  const t = createTranslator();

  function valueOrDash(value: string | number | null | undefined): string {
    if (value === null || value === undefined || value === '') return '-';
    return String(value);
  }

  function apparatLabel(device: FieldDevice): string {
    const parts = [device.apparat?.short_name, device.apparat?.name].filter(Boolean);
    return parts.length > 0 ? parts.join(' · ') : valueOrDash(device.apparat_id);
  }

  function systemPartLabel(device: FieldDevice): string {
    const parts = [device.system_part?.short_name, device.system_part?.name].filter(Boolean);
    return parts.length > 0 ? parts.join(' · ') : valueOrDash(device.system_part_id);
  }
</script>

<Card.Root class="border-border/70 bg-card/80">
  <Card.Header>
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1.5">
        <Card.Title class="text-xl">
          {$t('facility.sps_controller_system_type_detail.field_devices_title')}
        </Card.Title>
        <Card.Description>
          {$t('facility.sps_controller_system_type_detail.field_devices_desc')}
        </Card.Description>
      </div>

      <Badge variant={total > 0 ? 'secondary' : 'outline'}>{total}</Badge>
    </div>
  </Card.Header>

  <Card.Content>
    {#if fieldDevices.length === 0}
      <div
        class="rounded-md border border-dashed border-border px-4 py-10 text-center text-sm text-muted-foreground"
      >
        {$t('facility.sps_controller_system_type_detail.no_field_devices')}
      </div>
    {:else}
      <div class="overflow-hidden rounded-lg border border-border/70">
        <Table.Root>
          <Table.Header class="bg-muted/20">
            <Table.Row>
              <Table.Head>{$t('field_device.table.bmk')}</Table.Head>
              <Table.Head>{$t('field_device.table.text_fix')}</Table.Head>
              <Table.Head class="w-28">{$t('field_device.table.apparat_nr')}</Table.Head>
              <Table.Head>{$t('field_device.table.apparat')}</Table.Head>
              <Table.Head>{$t('field_device.table.system_part')}</Table.Head>
              <Table.Head class="w-32 text-right">{$t('facility.specifications')}</Table.Head>
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {#each fieldDevices as device (device.id)}
              <Table.Row>
                <Table.Cell>
                  <div class="space-y-1">
                    <div class="font-medium text-foreground">{valueOrDash(device.bmk)}</div>
                    {#if device.description}
                      <div class="line-clamp-2 text-xs text-muted-foreground">
                        {device.description}
                      </div>
                    {/if}
                  </div>
                </Table.Cell>
                <Table.Cell class="max-w-72">
                  <span class="line-clamp-2">{valueOrDash(device.text_fix)}</span>
                </Table.Cell>
                <Table.Cell class="font-mono text-sm">{valueOrDash(device.apparat_nr)}</Table.Cell>
                <Table.Cell>{apparatLabel(device)}</Table.Cell>
                <Table.Cell>{systemPartLabel(device)}</Table.Cell>
                <Table.Cell class="text-right">
                  <Badge variant={device.specification_id ? 'success' : 'outline'}>
                    {$t(
                      device.specification_id
                        ? 'field_device.table.spec_available'
                        : 'field_device.table.spec_missing'
                    )}
                  </Badge>
                </Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      </div>
      {#if total > fieldDevices.length}
        <p class="mt-2 text-xs text-muted-foreground">
          {$t('facility.sps_controller_system_type_detail.field_devices_limited', {
            shown: fieldDevices.length,
            total
          })}
        </p>
      {/if}
    {/if}
  </Card.Content>
</Card.Root>
